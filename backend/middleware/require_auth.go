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
		APIToken := c.GetHeader("X-API-Token")
		if APIToken != "" {
			permissions, err := auth.ValidateAPIToken(APIToken)
			if err != nil {
				c.Set("permissions", permissions)
				c.Next()
				return
			}
			log.Println("validating API token: failed with error %w", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.StatusBadRequestResponse{Error: "Invalid API Token"})
			return
		}

		JWT, err := c.Cookie("token")
		if err != nil || JWT == "" {
			log.Println("No JWT token provided with request to protected endpoint")
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.StatusBadRequestResponse{Error: "Missing Authorization Token"})
			return
		}

		claims, err := auth.ValidateJWT(JWT)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.StatusBadRequestResponse{Error: "Invalid token: " + err.Error()})
			return
		}
		if claims.Subject == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "Invalid token: missing user ID"})
			return
		}
		c.Set("userID", claims.Subject)
		c.Set("permissions", claims.Permissions)
		c.Next()
	}
}
