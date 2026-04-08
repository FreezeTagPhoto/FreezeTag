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

// @summary     Revoke API token
// @description Revoke an API token, preventing it from being used for authentication. Users can only revoke their own tokens
// @tags        tokens
// @router      /tokens/revoke/{id} [post]
// @param       id path string true "ID of the token to revoke"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     401 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
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

// @summary     Create API token
// @description Create a new API token for the authenticated user with the specified permissions and expiration time
// @tags        tokens
// @router      /tokens/create [post]
// @param       label query string false "Optional label for the token"
// @param       expiresAt query string false "Optional expiration time for the token in RFC3339 format (e.g. 2024-01-01T00:00:00Z)"
// @param       permission query []string true "List of permissions to assign to the token" collectionFormat(multi)
// @success     200 {object} services.ApiCreateToken
// @failure     400 {object} api.BadRequestResponse
// @failure     401 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
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

// @summary     Admin delete API token
// @description Permanently delete an API token from the database. This action cannot be undone. Admins can delete any token.
// @tags        tokens
// @router      /tokens/admin/delete/{id} [delete]
// @param       id path string true "ID of the token to delete"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (te TokenEndpoint) AdminDeleteUserToken(c *gin.Context) {
	tokenIDRaw := c.Param("id")
	tokenID, err := api.ParseParamIntoID[database.TokenID](tokenIDRaw)
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

// @summary     Admin revoke API token
// @description Revoke an API token, preventing it from being used for authentication. This is a softer action than deleting a token. Admins can revoke any token.
// @tags        tokens
// @router      /tokens/admin/revoke/{id} [post]
// @param       id path string true "ID of the token to revoke"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (te TokenEndpoint) AdminRevokeToken(c *gin.Context) {
	tokenIDRaw := c.Param("id")
	tokenID, err := api.ParseParamIntoID[database.TokenID](tokenIDRaw)
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
