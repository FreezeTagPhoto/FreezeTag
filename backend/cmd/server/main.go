//go:build !test

package main

import (
	"freezetag/backend/api/albums"
	"freezetag/backend/api/files"
	"freezetag/backend/api/jobs"
	"freezetag/backend/api/login"
	"freezetag/backend/api/logout"
	"freezetag/backend/api/metadata"
	"freezetag/backend/api/password"
	"freezetag/backend/api/permissions"
	"freezetag/backend/api/plugins"
	"freezetag/backend/api/search"
	"freezetag/backend/api/tags"
	"freezetag/backend/api/thumbnails"
	"freezetag/backend/api/tokens"
	"freezetag/backend/api/upload"
	"freezetag/backend/api/user"
	"freezetag/backend/middleware"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/images"
	"freezetag/backend/pkg/images/formats"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"

	"log"
	"os"
	"path"

	docs "freezetag/backend/cmd/docs"

	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	defaultDataDir = "./data"
)

type dependencies struct {
	imageRepository repositories.ImageRepository
	jobRepository   repositories.JobRepository

	jobService    services.JobService
	authService   services.AuthService
	pluginService services.PluginService

	albumRepository database.AlbumDatabase
}

// @title FreezeTag API
// @version 0.1
// @description This is the API access for the backend of the FreezeTag app.
//
// @basePath /
func main() {
	godotenv.Load(path.Join(defaultDataDir, ".env")) //nolint:errcheck
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"
	deps := initializeDependencies()

	RegisterEndpoints(router, deps)
	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerfiles.Handler)(c)
	})
	router.Run("0.0.0.0:3824") //nolint:errcheck
}

func initParserCollection() images.Parser {
	parserCollection := images.InitParserCollection()
	if err := parserCollection.RegisterParserFunc("*", formats.ParseBasic); err != nil {
		log.Fatalf("[ERR]  failed to register default parser: %v", err)
	}
	return parserCollection
}

func initializeDependencies() *dependencies {
	parserCollection := initParserCollection()

	jobRepo := repositories.NewDefaultJobRepository()
	manager := initDefaultManager(defaultDataDir, parserCollection)
	imageService := initDefaultImageRepository(manager, parserCollection)
	userRepo := manager.UserDB

	pluginService, err := services.InitDefaultPluginService("./plugins", imageService)
	if err != nil {
		log.Fatalf("[ERR]  error launching plugin service: %v", err)
	}
	log.Printf("[INFO] loaded plugins:")
	for _, plug := range pluginService.Plugins() {
		dis := ""
		if !plug.Enabled {
			dis = " [disabled]"
		}
		log.Printf("       - %s version %s%s", plug.Name, plug.Version, dis)
	}
	jobService := services.InitDefaultJobService(jobRepo, imageService, pluginService)
	authService := services.InitDefaultAuthService(userRepo, parserCollection)
	err = authService.EnsureLogin()
	if err != nil {
		log.Fatalf("[ERR]  error ensuring that the user can log in: %v", err)
	}

	return &dependencies{
		imageRepository: imageService,
		jobRepository:   jobRepo,
		jobService:      jobService,
		authService:     authService,
		pluginService:   pluginService,
		albumRepository: manager.AlbumDB,
	}
}

func initDefaultImageRepository(mgr *database.Manager, parserCollection images.Parser) repositories.ImageRepository {
	return repositories.InitImageRepository(mgr.ImageDB, parserCollection, path.Join(defaultDataDir, "images"))
}

func initDefaultManager(dataDir string, parserCollection images.Parser) *database.Manager {
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create data directory")
	}
	manager, err := database.NewDefaultManager(path.Join(dataDir, "database.db"))
	if err != nil {
		log.Fatalf("failed to initialize database manager: %v", err.Error())
	}
	return manager
}

// In order to make permission changes easy and keep unit tests modular, all middleware should
// be registered in here, and not other files.
func RegisterEndpoints(router *gin.Engine, deps *dependencies) {

	authGroup := router.Group("/")
	authGroup.Use(middleware.RequireAuth(deps.authService))

	initLoginEndpoints(router, deps)

	initApiKeyEndpoints(authGroup, deps)
	initPermissionsEndpoints(authGroup)
	initLogoutEndpoints(authGroup, deps)
	initPasswordEndpoints(authGroup, deps)
	initTagEndpoints(authGroup, deps)
	initUploadEndpoints(authGroup, deps)
	initThumbnailEndpoints(authGroup, deps)
	initSearchEndpoints(authGroup, deps)
	initMetadataEndpoints(authGroup, deps)
	initJobsEndpoints(authGroup, deps)
	initFileEndpoints(authGroup, deps)
	initUserEndpoints(authGroup, deps)
	initPluginEndpoints(authGroup, deps)
	initAlbumEndpoints(authGroup, deps)
}

func initAlbumEndpoints(baseGroup gin.IRouter, deps *dependencies) {
    ae := albums.InitAlbumEndpoint(deps.albumRepository)
    
    albumGroup := baseGroup.Group("/album")
    {
        albumGroup.GET("", ae.ListAlbums)
        albumGroup.POST("", ae.CreateAlbum)
        albumGroup.GET("/image/:id", ae.ListImageAlbums) 

        singleAlbum := albumGroup.Group("/:id")
        {
            singleAlbum.DELETE("", ae.DeleteAlbum)
			singleAlbum.GET("", ae.GetAlbumInfo)
            singleAlbum.PATCH("/name", ae.RenameAlbum)
            singleAlbum.PATCH("/visibility", ae.SetAlbumVisibility)
            singleAlbum.PUT("/permissions", ae.SetUserAlbumPermission)

			singleAlbum.POST("/images", ae.AddImageToAlbum)
            singleAlbum.GET("/images", ae.ListAlbumImages)
			singleAlbum.DELETE("/images/:image_id", ae.RemoveImageFromAlbum)
        }
    }
}

func initApiKeyEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	apiKeyGroup := baseGroup.Group("/tokens")
	{
		te := tokens.InitTokenEndpoint(deps.authService)
		apiKeyGroup.POST("/revoke/:id", middleware.RequirePermission(data.WriteToken), te.RevokeUserToken)
		apiKeyGroup.POST("/create", middleware.RequirePermission(data.WriteToken), te.CreateUserToken)
		apiKeyGroup.POST("/admin/revoke/:id", middleware.RequirePermission(data.WriteAnyToken), te.AdminRevokeToken)
		apiKeyGroup.DELETE("/admin/delete/:id", middleware.RequirePermission(data.WriteAnyToken), te.AdminDeleteUserToken)
	}
}

func initPermissionsEndpoints(baseGroup gin.IRouter) {
	permGroup := baseGroup.Group("/permissions")
	{
		pe := permissions.InitPermissionEndpoint()
		permGroup.GET("/list", middleware.RequirePermission(data.ReadPermissions), pe.ListPermissions)
	}
}

func initPluginEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	pluginGroup := baseGroup.Group("/plugins")
	{
		pe := plugins.InitPluginEndpoint(deps.pluginService, deps.jobService)
		pluginGroup.GET("/list", middleware.RequirePermission(data.ReadPlugins), pe.ListAll)
		pluginGroup.POST("/enable", middleware.RequirePermission(data.WritePlugins), pe.SetEnabled)
		pluginGroup.POST("/run", middleware.RequirePermission(data.WritePlugins), pe.RunManual)
		pluginGroup.GET("/config", middleware.RequirePermission(data.ReadPlugins), pe.ReadConfig)
		pluginGroup.POST("/config", middleware.RequirePermission(data.WritePlugins), pe.ChangeConfig)
	}
}

func initPasswordEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	passwordGroup := baseGroup.Group("/password")
	{
		pe := password.InitPasswordEndpoint(deps.authService)
		passwordGroup.POST("/change", middleware.RequirePermissionOrSelf(data.WriteUser), pe.ChangePassword)
	}
}

func initLoginEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	le := login.InitLoginEndpoint(deps.authService)
	baseGroup.POST("/login", le.Login)
	baseGroup.GET("/login", le.LoginInfo)
}

func initLogoutEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	lo := logout.InitLogoutEndpoint(deps.authService)
	baseGroup.POST("/logout", lo.HandleLogout)
}

func initUserEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	userGroup := baseGroup.Group("/users")
	{
		ue := user.InitUserEndpoint(deps.authService)
		userGroup.GET("/profile-picture/:id", middleware.RequirePermissionOrSelf(data.ReadUser), ue.GetProfilePicture)
		userGroup.POST("/profile-picture/:id", middleware.RequirePermissionOrSelf(data.WriteUser), ue.SetProfilePicture)
		userGroup.DELETE("/profile-picture/:id", middleware.RequirePermissionOrSelf(data.WriteUser), ue.ResetProfilePicture)

		userGroup.GET("/:id", middleware.RequirePermissionOrSelf(data.ReadUser), ue.GetUser)
		userGroup.DELETE("/:id", middleware.RequirePermission(data.WriteUser), ue.DeleteUser)

		userGroup.GET("/all", middleware.RequirePermission(data.ReadUser), ue.ListUsers)
		userGroup.POST("/create", middleware.RequirePermission(data.WriteUser), ue.CreateUser)

		userGroup.GET("/permissions/:id", middleware.RequirePermission(data.ReadPermissions), ue.GetPermissions)
		userGroup.POST("/permissions/:id", middleware.RequirePermission(data.WritePermissions), ue.AddPermissions)
		userGroup.DELETE("/permissions/:id", middleware.RequirePermission(data.WritePermissions), ue.RevokePermissions)

		userGroup.POST("/visibility/:id", middleware.RequirePermissionOrSelf(data.WriteUser), ue.SetUserVisibilityMode)
	}
}

func initFileEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	fileGroup := baseGroup.Group("/file")
	{
		fe := files.InitFileEndpoint(deps.imageRepository)
		fileGroup.GET("/download/:id", middleware.RequirePermission(data.ReadFiles), fe.HandleGet)
		fileGroup.DELETE("/delete/:id", middleware.RequirePermission(data.WriteFiles), fe.HandleDelete)
	}
}

func initJobsEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	jobGroup := baseGroup.Group("/jobs")
	{
		je := jobs.InitJobsEndpoint(deps.jobService)
		jobGroup.GET("/details/:id", je.Details)
		jobGroup.GET("/summary/:id", je.Summary)
		jobGroup.GET("/list", je.List)
		jobGroup.POST("/cancel/:id", middleware.RequirePermission(data.WriteJobs), je.Cancel)
	}
}

func initMetadataEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	metadataGroup := baseGroup.Group("/metadata")
	{
		me := metadata.InitMetadataEndpoint(deps.imageRepository)
		metadataGroup.GET("/:id", middleware.RequirePermission(data.ReadFiles), me.Metadata)
	}
}

func initSearchEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	searchGroup := baseGroup.Group("/search")
	{
		se := search.InitSearchEndpoint(deps.imageRepository)
		searchGroup.GET("", middleware.RequirePermission(data.ReadFiles), se.Search)
	}
}

func initThumbnailEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	thumbnailGroup := baseGroup.Group("/thumbnails")
	{
		te := thumbnails.InitThumbnailEndpoint(deps.imageRepository)
		thumbnailGroup.GET("/:id", middleware.RequirePermission(data.ReadFiles), te.HandleGet)
	}
}

func initUploadEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	uploadGroup := baseGroup.Group("/upload")
	{
		ue := upload.InitUploadEndpoint(deps.jobService)
		uploadGroup.POST("", middleware.RequirePermission(data.WriteFiles), ue.Upload)
	}
}

func initTagEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	tagGroup := baseGroup.Group("/tag")
	{
		te := tags.InitTagEndpoint(deps.imageRepository)
		tagGroup.DELETE("/remove", middleware.RequirePermission(data.WriteTags), te.HandleDelete)
		tagGroup.DELETE("/delete", middleware.RequirePermission(data.WriteTags), te.HandleDeleteFull)
		tagGroup.POST("/add", middleware.RequirePermission(data.WriteTags), te.HandlePost)

		tagGroup.GET("/list", middleware.RequirePermission(data.ReadTags), te.ListTags)
		tagGroup.GET("/list/:id", middleware.RequirePermission(data.ReadTags), te.ImageTags)
		tagGroup.GET("/counts", middleware.RequirePermission(data.ReadTags), te.ListCounts)
		tagGroup.GET("/search", middleware.RequirePermission(data.ReadTags), te.ListCountsQuery)
	}
}
