//go:build !test

package main

import (
	"freezetag/backend/api/upload"
	"freezetag/backend/pkg/repositories"

	"github.com/gin-gonic/gin"
)

type repository struct{}

func (repository) StoreImageBytes(data []byte, filename string) (repositories.Result) {
	return repositories.Result{
		Success: &repositories.ImageHandleSuccess{Id: 67, Filename: filename},
		Err:     nil,
	}		
}

func (repository) RetrieveImage(uid uint) (any, error) {
	return nil, nil
}

func main() {
	router := gin.Default()
	var repo repository
	upload.InitUploadEndpoint(repo).RegisterEndpoints(router)
	router.Run("localhost:3824") //nolint:errcheck // no need to check return value
}
