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

func (le LoginEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.POST("/login", le.HandleLogin)
}

// @summary Authenticate user and return a token
// @description Authenticates a user using form parameters "username" and "password". On success returns a JSON payload containing an authentication token.
// @tags auth, login
// @accept application/x-www-form-urlencoded
// @produce application/json
// @param username formData string true "Username"
// @param password formData string true "Password"
// @success 200 {object} api.StatusLoginSuccess "Authentication successful"
// @failure 401 {object} api.StatusLoginFail "Authentication failed"
// @router /login [post]
func (le LoginEndpoint) HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	
	token, err := le.authService.AuthenticateUser(username, password)	
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.StatusLoginFail{Error: "authentication failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.StatusLoginSuccess{Token: token})
}