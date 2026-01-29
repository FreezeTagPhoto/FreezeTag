package login

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginEndpoint struct {
	authService services.AuthService
}

func InitLoginEndpoint(authService services.AuthService) LoginEndpoint {
	return LoginEndpoint{
		authService: authService,
	}
}

func (le LoginEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("/login", le.HandleLogin)
}

// @summary Authenticate user and return a token
// @description Authenticates a user using form parameters "username" and "password". On success returns a JSON payload containing an authentication token.
// @tags auth, login
// @accept application/json
// @produce application/json
// @param request body api.LoginCredentials true "User Login Details"
// @success 200 {object} api.StatusLoginSuccess "Authentication successful"
// @failure 401 {object} api.StatusLoginFail "Authentication failed"
// @router /login [post]
func (le LoginEndpoint) HandleLogin(c *gin.Context) {
	var req api.LoginCredentials
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "invalid request: " + err.Error()})
		return
	}

	token, err := le.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.StatusLoginFail{Error: "authentication failed: " + err.Error()})
		return
	}
	c.SetCookieData(&http.Cookie{
		Name:     "token",
		Value:    token,
		Secure:   false,
		HttpOnly: true,
	})
	c.JSON(http.StatusOK, api.StatusLoginSuccess{Token: token})
}
