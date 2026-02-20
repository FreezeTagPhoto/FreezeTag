package tokens

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TokenEndpoint struct {
	authService services.AuthService
}

func InitTokenEndpoint(auth services.AuthService) TokenEndpoint {
	return TokenEndpoint{
		authService: auth,
	}
}

func (te TokenEndpoint) RevokeUserToken(c *gin.Context) {
	tokenIDRaw := c.Param("id")
	tokenID, err := api.ParseParamIntoID[database.TokenID](tokenIDRaw)
	userIDRaw, exists := c.Get("userID")
	if !exists {
        c.JSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "Unauthorized"})
        return
    }
	userID, err := api.ParseParamIntoID[database.UserID](userIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid user ID"})
		return
	}
	err = te.authService.RevokeAPIToken(userID, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to revoke token"})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Token revoked successfully"})
}

func (te TokenEndpoint) AdminDeleteUserToken(c *gin.Context) {
	tokenIdRaw := c.Param("id")
	tokenID, err := api.ParseParamIntoID[database.TokenID](tokenIdRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid token ID"})
		return
	}
	err = te.authService.DeleteAPIToken(tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to delete token"})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Token deleted successfully"})

}

func (te TokenEndpoint) AdminRevokeToken(c *gin.Context) {
	tokenIdRaw := c.Param("id")
	tokenID, err := api.ParseParamIntoID[database.TokenID](tokenIdRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid token ID"})
		return
	}
	err = te.authService.AdminRevokeAPIToken(tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to revoke token"})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Token revoked successfully"})
}
