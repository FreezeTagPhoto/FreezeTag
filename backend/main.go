//go:build !test

package main

import (
	"freezetag/backend/api/test"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	test.RegisterEndpoints(router)
	router.Run("localhost:8080") //nolint:errcheck // no need to check return value
}
