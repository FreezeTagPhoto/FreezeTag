package api

import "github.com/gin-gonic/gin"

type ApiEndpoint interface { 
	RegisterEndpoints(e *gin.Engine) 
}  