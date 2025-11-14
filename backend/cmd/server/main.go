//go:build !test

package main

import (
	"freezetag/backend/api/search"
	"freezetag/backend/api/thumbnails"
	"freezetag/backend/api/upload"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images"
	"freezetag/backend/pkg/images/formats"
	"freezetag/backend/pkg/repositories"
	"log"

	"github.com/gin-gonic/gin"
)

const defaultImageFolder = "./images/"

func main() {
	router := gin.Default()
	repo := initRepository(defaultImageFolder)
	RegisterEndpoints(router, repo)
	router.Run("localhost:3824") //nolint:errcheck // no need to check return value
}

func initParserCollection() images.Parser {
	parserCollection := images.InitParserCollection()
	if err := parserCollection.RegisterParserFunc("*.{cr3,nef,dng,hei{c,f}}", formats.ParseRaw); err != nil {
		log.Fatalf("[ERROR] failed to register RAW parser: %v", err)
	}
	if err := parserCollection.RegisterParserFunc("*.{png,jpg,jpeg}", formats.ParseBasic); err != nil {
		log.Fatalf("[ERROR] failed to register basic parser: %v", err)
	}
	return parserCollection
}

// initRepository initializes the image repository with a SQLite database and an image parser collection.
func initRepository(imageFolder string) repositories.ImageRepository {
	db, err := database.InitSQLiteImageDatabase("database.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err.Error())
	}
	parserCollection := initParserCollection()
	return repositories.InitImageRepository(db, parserCollection, imageFolder)
}

func RegisterEndpoints(router *gin.Engine, repo repositories.ImageRepository) {
	upload.InitUploadEndpoint(repo).RegisterEndpoints(router)
	thumbnails.InitThumbnailEndpoint(repo).RegisterEndpoints(router)
	search.InitSearchEndpoint(repo).RegisterEndpoints(router)
	// Other endpoints would be registered here
}
