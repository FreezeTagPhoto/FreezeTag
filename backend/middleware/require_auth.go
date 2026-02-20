package middleware

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(auth services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		api_key := c.GetHeader("X-API-Key")
		if api_key != "" {
			claims, err := auth.ValidateAPIToken(api_key)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "Invalid API Key"})
				return
			}
			c.Set("userID", claims.UserID)
			c.Set("permissions", claims.Permissions)
			c.Next()
			return
		}

		JWT, err := c.Cookie("token")
		if err != nil || JWT == "" {
			log.Println("No JWT token provided with request to protected endpoint")
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "Missing Authorization Token"})
			return
		}

		claims, err := auth.ValidateJWT(JWT)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "Invalid token: " + err.Error()})
			return
		}
		if claims.Subject == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid token: missing user ID"})
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("permissions", claims.Permissions)
		c.Next()
	}
}
