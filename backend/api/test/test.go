package test

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTestEndpoint(c *gin.Context) {
	c.String(http.StatusOK, "Test Successful")
}

func RegisterEndpoints(e *gin.Engine) {
	e.GET("/test", GetTestEndpoint)
}
