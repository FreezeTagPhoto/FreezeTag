//go:build !test

package main

import (
	"freezetag/backend/api/jobquery"
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

// @title FreezeTag API
// @version 0.1
// @description This is the API access for the backend of the FreezeTag app.
//
// @basePath /
func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"
	imageRepo := initDefaultImageRepository(defaultImageFolder)
	jobRepo := initDefaultJobRepository()
	jobService := services.InitDefaultJobService(jobRepo, imageRepo)
	
	RegisterEndpoints(router, imageRepo, jobRepo, jobService)
	router.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "" || c.Param("any") == "/" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerfiles.Handler)(c)
	})
	router.Run("localhost:3824") //nolint:errcheck // no need to check return value
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

// initDefaultImageRepository initializes the image repository with a SQLite database and an image parser collection.
func initDefaultImageRepository(imageFolder string) repositories.ImageRepository {
	db, err := database.InitSQLiteImageDatabase("database.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err.Error())
	}
	parserCollection := initParserCollection()
	return repositories.InitImageRepository(db, parserCollection, imageFolder)
}

func initDefaultJobRepository() repositories.JobRepository {
	return repositories.NewDefaultJobRepository()
}

func RegisterEndpoints(router *gin.Engine, repo repositories.ImageRepository, jobRepo repositories.JobRepository, jobService services.JobService) {
	upload.InitUploadEndpoint(jobService).RegisterEndpoints(router)	

	thumbnails.InitThumbnailEndpoint(repo).RegisterEndpoints(router)
	search.InitSearchEndpoint(repo).RegisterEndpoints(router)
	tags.InitTagEndpoint(repo).RegisterEndpoints(router)

	jobquery.InitJobQueryEndpoint(jobRepo).RegisterEndpoints(router)
}
