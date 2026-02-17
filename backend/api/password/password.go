package password

import (
	"freezetag/backend/pkg/services"

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

func (pe PasswordEndpoint) ChangePassword(c *gin.Context) {
	// id, err := api.GetUserIDFromString(c.Param("id"))
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: err.Error()})
	// 	return
	// }
	// var req api.PasswordChangeRequest
	// if err := c.ShouldBind(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "invalid request: " + err.Error()})
	// 	return
	// }

	// err = pe.authService.ChangeUserawPassword(id, req.NewPassword)
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, api.StatusUnauthorizedResponse{Error: "invalid credentials"})
	// 	return
	// }
	// c.JSON(http.StatusOK, api.StatusOkResponse{Message: "password changed successfully"})
}