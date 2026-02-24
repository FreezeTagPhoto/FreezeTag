package password

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordEndpoint struct {
	authService services.AuthService
}

func InitPasswordEndpoint(authService services.AuthService) PasswordEndpoint {
	return PasswordEndpoint{
		authService: authService,
	}
}

// @summary Change user password
// @description Allows an authenticated user to change their password by providing their current password and a new password.
// @tags auth, password
// @accept application/json
// @produce application/json
// @param request body api.PasswordChangeRequest true "Password Change Request"
// @success 200 {object} api.MessageResponse "Password changed successfully"
// @failure 400 {object} api.BadRequestResponse "Invalid request"
// @failure 401 {object} api.BadRequestResponse "User not authenticated"
// @failure 500 {object} api.BadRequestResponse "Failed to change password"
// @router /password/change [post]
func (pe PasswordEndpoint) ChangePassword(c *gin.Context) {
	var req api.PasswordChangeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request: " + err.Error()})
		return
	}
	sub, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "user ID not found in jwt token"})
		return
	}
	subString, ok := sub.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "invalid user ID in JWT token"})
		return
	}
	userID, err := api.ParseParamIntoID[database.UserID](subString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.BadRequestResponse{Error: "invalid user ID in JWT token"})
		return
	}
	if err := pe.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, api.BadRequestResponse{Error: "failed to change password: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "password changed successfully"})
}
