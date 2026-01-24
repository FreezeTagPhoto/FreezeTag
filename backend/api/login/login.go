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