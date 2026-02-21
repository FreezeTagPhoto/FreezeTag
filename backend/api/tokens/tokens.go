package tokens

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateTokenRequest struct {
	Label     string     `form:"label"`
	ExpiresAt *time.Time `form:"expiresAt" time_format:"2006-01-02T15:04:05Z07:00"` // RFC3339
}

type TokenEndpoint struct {
	authService services.AuthService
}

func InitTokenEndpoint(auth services.AuthService) TokenEndpoint {
	return TokenEndpoint{
		authService: auth,
	}
}

func (te TokenEndpoint) RevokeUserToken(c *gin.Context) {

	// parse the token ID from the URL parameter
	tokenIDRaw := c.Param("id")
	tokenID, err := api.ParseParamIntoID[database.TokenID](tokenIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid token ID"})
		return
	}

	// get the user ID from the JWT token
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

	// attempt to revoke the token
	err = te.authService.RevokeAPIToken(userID, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to revoke token"})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Token revoked successfully"})
}


func (te TokenEndpoint) CreateUserToken(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "Unauthorized"})
		return
	}
	permissions, err := api.QueryPermissionsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	userID, err := api.ParseParamIntoID[database.UserID](userIDRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid user ID"})
		return
	}

	var req CreateTokenRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid query parameters"})
		return
	}

	token, err := te.authService.CreateAPIToken(userID, permissions, req.ExpiresAt, req.Label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to create token"})
		return
	}
	c.JSON(http.StatusOK, token)
}


// Admin versions of the above requests. Admins can operate on any token, not just the tokens assigned to them

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
