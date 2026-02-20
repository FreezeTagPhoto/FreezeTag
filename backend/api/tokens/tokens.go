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
	// ADMIN should be the only one who can delete tokens, but users should be able to revoke their own tokens. 
	// Revoke is a soft delete that prevents the token from being used but keeps the record for auditing purposes

}

func (te TokenEndpoint) AdminRevokeToken(c *gin.Context) {
	// 
}