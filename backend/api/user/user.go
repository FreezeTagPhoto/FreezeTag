package user

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserEndpoint struct {
	userRepo repositories.UserRepository
}

func InitUserEndpoint(userRepo repositories.UserRepository) UserEndpoint {
	return UserEndpoint{userRepo: userRepo}
}

func (ue UserEndpoint) RegisterEndpoints(router gin.IRoutes) {
	router.GET("/user/:id", ue.GetUser)
	router.GET("/users", ue.ListUsers)
}

// @Summary      Get a public user by ID
// @Description  Retrieves a user's public information by their ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  database.PublicUser
// @Failure      400  {object}  api.StatusBadRequestResponse
// @Failure      404  {object}  api.StatusBadRequestResponse
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

func (ue UserEndpoint) ListUsers(c *gin.Context) {
	users, err := ue.userRepo.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.StatusBadRequestResponse{Error: "Failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
