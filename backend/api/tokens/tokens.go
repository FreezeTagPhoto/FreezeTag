package tokens

import (
	"freezetag/backend/pkg/services"

	"github.com/gin-gonic/gin"
)

type TokenEndpoint struct {
	authService services.AuthService
}

func InitTokenEndpoint(auth services.AuthService) TokenEndpoint {
	return TokenEndpoint{auth}
}

func (te TokenEndpoint) RevokeToken(c *gin.Context) {

}
