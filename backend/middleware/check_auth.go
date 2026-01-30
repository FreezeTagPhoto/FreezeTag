package middleware

import (
	"freezetag/backend/pkg/services"

	"github.com/gin-gonic/gin"
)

func CheckAuth(auth services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		JWT, err := c.Cookie("token")
		if err != nil || JWT == "" {
			c.Set("authenticated", false)
			c.Next()
			return
		}

		_, err = auth.ValidateJWT(JWT)
		if err != nil {
			c.Set("authenticated", false)
			c.Next()
			return
		}
		c.Set("authenticated", true)
		c.Next()
	}
}
