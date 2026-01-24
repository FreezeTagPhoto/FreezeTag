package createuser

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateUserEndpoint struct {
	authService services.AuthService
}


// Creates a new CreateUserEndpoint with the given auth service.
func InitCreateUserEndpoint(authService services.AuthService) CreateUserEndpoint {
	return CreateUserEndpoint{
		authService: authService,
	}
}

// Registers the job query endpoints to the given Gin engine.
func (ce CreateUserEndpoint) RegisterEndpoints(e *gin.Engine) {
	e.POST("/createuser", je.HandleCreateUser)
}

// @summary     Create a new user
// @description Creates a new user with the given username and password.
// @tags auth, users
// @accept application/x-www-form-urlencoded
// @produce application/json
// @param username formData string true "Username"
// @param password formData string true "Password"
// @success 200 {object} database.PublicUser "User created successfully"
// @failure 400 {object} api.StatusBadRequestResponse "Failed to create user"
// @router /createuser [post]
func (ce CreateUserEndpoint) HandleCreateUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	user, err := ce.authService.AddUser(username, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "failed to create user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)

	
}