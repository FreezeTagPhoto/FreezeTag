//go:build !test

package main

import (
	"freezetag/backend/api/jobquery"
	"freezetag/backend/api/login"
	"freezetag/backend/api/metadata"
	"freezetag/backend/api/search"
	"freezetag/backend/api/tags"
	"freezetag/backend/api/thumbnails"
	"freezetag/backend/api/upload"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images"
	"freezetag/backend/pkg/images/formats"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"log"

	docs "freezetag/backend/cmd/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"net/http"

	"github.com/gin-gonic/gin"
)

const defaultImageFolder = "./images"

type dependencies struct {
	imageRepository repositories.ImageRepository
	jobRepository   repositories.JobRepository
	userRepository  repositories.UserRepository
	
	jobService      services.JobService
	authService     services.AuthService
}

// @title FreezeTag API
// @version 0.1
// @description This is the API access for the backend of the FreezeTag app.
//
// @basePath /
func main() {
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
	imageRepo := initDefaultImageRepository(defaultImageFolder)
	userRepo := initDefaultUserRepository()

	jobService := services.InitDefaultJobService(jobRepo, imageRepo)
	authService := services.InitDefaultAuthService(userRepo)

	return &dependencies{
		imageRepository: imageRepo,
		jobRepository:   jobRepo,
		userRepository:  userRepo,
		jobService:      jobService,
		authService:     authService,
	}
}

func initDefaultUserRepository() repositories.UserRepository {
	db, err := database.InitSQLiteUserDatabase("users.db")
	if err != nil {
		log.Fatalf("failed to initialize user database: %v", err.Error())
	}
	return repositories.InitDefaultUserRepository(db)
}

func initDefaultImageRepository(imageFolder string) repositories.ImageRepository {
	db, err := database.InitSQLiteImageDatabase("database.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err.Error())
	}
	parserCollection := initParserCollection()
	return repositories.InitImageRepository(db, parserCollection, imageFolder)
}

func RegisterEndpoints(router *gin.Engine, deps *dependencies) {
	upload.InitUploadEndpoint(deps.jobService).RegisterEndpoints(router)
	login.InitLoginEndpoint(deps.authService).RegisterEndpoints(router)

	thumbnails.InitThumbnailEndpoint(deps.imageRepository).RegisterEndpoints(router)
	search.InitSearchEndpoint(deps.imageRepository).RegisterEndpoints(router)
	tags.InitTagEndpoint(deps.imageRepository).RegisterEndpoints(router)
	metadata.InitMetadataEndpoint(deps.imageRepository).RegisterEndpoints(router)
	jobquery.InitJobQueryEndpoint(deps.jobRepository).RegisterEndpoints(router)
}
