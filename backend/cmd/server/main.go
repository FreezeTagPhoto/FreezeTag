//go:build !test

package main

import (
	"freezetag/backend/api/files"
	"freezetag/backend/api/jobs"
	"freezetag/backend/api/login"
	"freezetag/backend/api/logout"
	"freezetag/backend/api/metadata"
	"freezetag/backend/api/search"
	"freezetag/backend/api/tags"
	"freezetag/backend/api/thumbnails"
	"freezetag/backend/api/password"
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
	"strings"

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
	userRepository  repositories.UserRepository

	jobService  services.JobService
	authService services.AuthService
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
	if err := parserCollection.RegisterParserFunc("*.{cr3,nef,dng,hei{c,f}}", formats.ParseRaw); err != nil {
		log.Fatalf("[ERROR] failed to register RAW parser: %v", err)
	}
	if err := parserCollection.RegisterParserFunc("*.{png,jpg,jpeg,webp}", formats.ParseBasic); err != nil {
		log.Fatalf("[ERROR] failed to register basic parser: %v", err)
	}
	return parserCollection
}

func initializeDependencies() *dependencies {
	jobRepo := repositories.NewDefaultJobRepository()
	imageRepo := initDefaultImageRepository(defaultDataDir)
	userRepo := initDefaultUserRepository(defaultDataDir)

	pluginService, err := services.InitDefaultPluginService("./plugins", imageRepo)
	if err != nil {
		log.Fatalf("[ERR]  error launching plugin service: %v", err)
	}
	log.Printf("[INFO] loaded plugins:")
	plugs := strings.Join(pluginService.AllPlugins(), ", ")
	if plugs == "" {
		log.Printf("[INFO] no plugins loaded")
	} else {
		log.Printf("[INFO] %s", plugs)
	}
	jobService := services.InitDefaultJobService(jobRepo, imageRepo, pluginService)
	authService := services.InitDefaultAuthService(userRepo)
	err = authService.EnsureLogin()
	if err != nil {
		log.Fatalf("[ERR]  error ensuring that the user can log in: %v", err)
	}

	return &dependencies{
		imageRepository: imageRepo,
		jobRepository:   jobRepo,
		userRepository:  userRepo,
		jobService:      jobService,
		authService:     authService,
	}
}

func initDefaultUserRepository(dataDir string) repositories.UserRepository {
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create data directory")
	}
	db, err := database.InitSQLiteUserDatabase(path.Join(dataDir, "users.db"))
	if err != nil {
		log.Fatalf("failed to initialize user database: %v", err.Error())
	}
	return repositories.InitDefaultUserRepository(db)
}

func initDefaultImageRepository(dataDir string) repositories.ImageRepository {
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create data directory")
	}
	db, err := database.InitSQLiteImageDatabase(path.Join(dataDir, "database.db"))
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err.Error())
	}
	parserCollection := initParserCollection()
	return repositories.InitImageRepository(db, parserCollection, path.Join(dataDir, "images"))
}

// In order to make permission changes easy and keep unit tests modular, all middleware should
// be registered in here, and not other files.
func RegisterEndpoints(router *gin.Engine, deps *dependencies) {

	authGroup := router.Group("/")
	authGroup.Use(middleware.RequireAuth(deps.authService))

	initLoginEndpoints(router, deps)
	
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
}

func initPasswordEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	passwordGroup := baseGroup.Group("/password")
	{
		pe := password.InitPasswordEndpoint(deps.authService)
		passwordGroup.POST("/change", pe.ChangePassword)
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
		ue := user.InitUserEndpoint(deps.userRepository, deps.authService)
		userGroup.GET("/:id", middleware.RequirePermission(data.ReadUser), ue.GetUser)
		userGroup.GET("/all", middleware.RequirePermission(data.ReadUser), ue.ListUsers)

		// eventually createuser needs to be just /user and then the userGroup can use "/user" as the base path,
		// but for now this wont cause merge conflicts
		userGroup.POST("/create", middleware.RequirePermission(data.CreateUser), ue.CreateUser)
		userGroup.POST("/permissions/:id", middleware.RequirePermission(data.WritePermissions), ue.AddPermissions)

		userGroup.DELETE("/permissions/:id", middleware.RequirePermission(data.WritePermissions), ue.RevokePermissions)
		userGroup.DELETE("/:id", middleware.RequirePermission(data.DeleteUser), ue.DeleteUser)

	}
}

func initFileEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	fileGroup := baseGroup.Group("/file")
	{
		fe := files.InitFileEndpoint(deps.imageRepository)
		fileGroup.GET("/:id", middleware.RequirePermission(data.ReadFiles), fe.HandleGet)
	}
}

func initJobsEndpoints(baseGroup gin.IRouter, deps *dependencies) {
	jobGroup := baseGroup.Group("/jobs")
	{
		je := jobs.InitJobsEndpoint(deps.jobService)
		jobGroup.GET("/details/:id", je.Details)
		jobGroup.GET("/summary/:id", je.Summary)
		jobGroup.GET("/list", je.List)
		jobGroup.POST("/cancel/:id", je.Cancel)
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
		tagGroup.POST("/add", middleware.RequirePermission(data.WriteTags), te.HandlePost)

		tagGroup.GET("/list", middleware.RequirePermission(data.ReadTags), te.ListTags)
		tagGroup.GET("/list/:id", middleware.RequirePermission(data.ReadTags), te.ImageTags)
		tagGroup.GET("/counts", middleware.RequirePermission(data.ReadTags), te.ListCounts)
	}
}
