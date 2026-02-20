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

func (te TokenEndpoint) RevokeUserToken(c *gin.Context) {

}

func (te TokenEndpoint) DeleteUserToken(c *gin.Context) {

}

func (te TokenEndpoint) AdminRevokeToken(c *gin.Context) {

}