package middleware

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireAuth(auth services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {

		JWT, err := c.Cookie("token")

		// Fallback to Authorization header if cookie is not present
		if err != nil || JWT == "" {
			JWT := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
			if JWT == "" {
				log.Println("No JWT token provided with request to protected endpoint")
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.StatusBadRequestResponse{Error: "Missing Authorization Token"})
				return
			}
		}

		claims, err := auth.ValidateJWT(JWT)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.StatusBadRequestResponse{Error: "Invalid token: " + err.Error()})
			return
		}
		c.Set("userID", claims["sub"])
		c.Next()
	}
}
