package createuser

import (
	"freezetag/backend/api"
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
func (ce CreateUserEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("/createuser", ce.HandleCreateUser)
}

// @summary      Create a new user
// @description  Creates a new user with the given username and password.
// @tags         auth, users
// @accept       json
// @produce      json
// @param        request body api.LoginCredentials true "User Registration Details"
// @success      200 {object} database.PublicUser "User created successfully"
// @failure      400 {object} api.StatusBadRequestResponse "Failed to create user"
// @router       /createuser [post]
func (ce CreateUserEndpoint) HandleCreateUser(c *gin.Context) {
	var req api.LoginCredentials
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "invalid request: " + err.Error()})
		return
	}
	user, err := ce.authService.AddUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "failed to create user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
