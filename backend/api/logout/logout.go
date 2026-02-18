package logout

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogoutEndpoint struct {
	authService services.AuthService
}

func InitLogoutEndpoint(auth services.AuthService) LogoutEndpoint {
	return LogoutEndpoint{auth}
}

// @summary invalidate the current user's session token
// @tags    auth, logout
// @success 200 {object} api.LogoutSuccessResponse
// @failure 401 {object} api.LoginFailResponse
// @router /logout [post]
func (lo LogoutEndpoint) HandleLogout(c *gin.Context) {
	c.SetCookieData(&http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	})
	c.JSON(http.StatusOK, api.LogoutSuccessResponse{Status: "ok"})
}
