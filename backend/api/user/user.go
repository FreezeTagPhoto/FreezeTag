package user

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserEndpoint struct {
	userRepo repositories.UserRepository
	authService services.AuthService
}

func InitUserEndpoint(userRepo repositories.UserRepository, authService services.AuthService) UserEndpoint {
	return UserEndpoint{userRepo: userRepo, authService: authService}
}

// @Summary      Get a public user by ID
// @Description  Retrieves a user's public information by their ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  database.PublicUser
// @Failure      400  {object}  api.StatusBadRequestResponse
// @Failure      500  {object}  api.StatusBadRequestResponse
// @Router       /user/{id} [get]
func (ue UserEndpoint) GetUser(c *gin.Context) {
	userIDString := c.Param("id")
	var id database.UserID
	if num, err := strconv.ParseInt(userIDString, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "Invalid user ID parameter: " + userIDString})
		return
	} else {
		id = database.UserID(num)
	}
	user, err := ue.userRepo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusBadRequestResponse{Error: "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary      List the users in the system
// @Description  Retrieves a list of all users in the system.
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   database.PublicUser
// @Failure      500  {object}  api.StatusBadRequestResponse
// @Router       /users [get]
func (ue UserEndpoint) ListUsers(c *gin.Context) {
	users, err := ue.userRepo.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusBadRequestResponse{Error: "Failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (ue UserEndpoint) DeleteUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, api.StatusBadRequestResponse{Error: "Delete user endpoint not implemented yet"})
}

func (ue UserEndpoint) CreateUser(c *gin.Context) {
	var req api.LoginCredentials
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "invalid request: " + err.Error()})
		return
	}
	user, err := ue.authService.AddUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "failed to create user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
