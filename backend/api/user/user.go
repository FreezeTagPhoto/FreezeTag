package user

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/repositories"
	"freezetag/backend/pkg/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserEndpoint struct {
	userRepo    repositories.UserRepository
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
// @Failure      400  {object}  api.BadRequestResponse
// @Failure      500  {object}  api.ServerErrorResponse
// @Router       /users/{id} [get]
func (ue UserEndpoint) GetUser(c *gin.Context) {
	userIDString := c.Param("id")
	var id database.UserID
	if num, err := strconv.ParseInt(userIDString, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid user ID parameter: " + userIDString})
		return
	} else {
		id = database.UserID(num)
	}
	user, err := ue.userRepo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "User not found"})
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
// @Failure      500  {object}  api.ServerErrorResponse
// @Router       /users/all [get]
func (ue UserEndpoint) ListUsers(c *gin.Context) {
	users, err := ue.userRepo.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// @Summary      Delete a user by ID
// @Description  Deletes a user from the system by their ID.
// @Tags         users
// @Accept       application/json
// @Produce      application/json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  api.MessageResponse
// @Failure      400  {object}  api.BadRequestResponse
// @Failure      500  {object}  api.ServerErrorResponse
// @Router       /users/{id} [delete]
func (ue UserEndpoint) DeleteUser(c *gin.Context) {
	userIDString := c.Param("id")
	id, err := api.GetUserIDFromString(userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	err = ue.userRepo.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: fmt.Sprintf("user %s deleted", userIDString)})
}

// @Summary      Create a new user
// @Description  Creates a new user with the provided username and password, granted they have permission to create users.
// @Tags         users
// @Accept       application/json
// @Produce      application/json
// @param 		 request body      api.LoginCredentials true "User Login Details"
// @Success      200     {object}  database.PublicUser
// @Failure      400     {object}  api.BadRequestResponse
// @Failure      500     {object}  api.ServerErrorResponse
// @Router       /users/create   [post]
func (ue UserEndpoint) CreateUser(c *gin.Context) {
	var req api.LoginCredentials
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request: " + err.Error()})
		return
	}
	user, err := ue.authService.AddUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "failed to create user: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary      Grant permissions for a user
// @Description  Grants specified permissions for a user by their ID.
// @Tags         users
// @Accept       application/json
// @Produce      application/json
// @Param        id   path      int  true  "User ID"
// @Param        permission query []string true "Permissions to grant" collectionFormat(multi)
// @Success      200  {object}  api.MessageResponse
// @Failure      400  {object}  api.BadRequestResponse
// @Failure      500  {object}  api.ServerErrorResponse
// @Router       /users/permissions/{id} [post]
func (ue UserEndpoint) AddPermissions(c *gin.Context) {
	userIDString := c.Param("id")
	id, err := api.GetUserIDFromString(userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	permissions, err := queryPermissionsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	err = ue.userRepo.GrantPermissions(id, permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "failed to grant permissions: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "permissions granted"})
}

// @Summary      Revoke permissions for a user
// @Description  Revokes specified permissions for a user by their ID.
// @Tags         users
// @Accept       application/json
// @Produce      application/json
// @Param        id   path      int  true  "User ID"
// @Param        permission query []string true "Permissions to revoke in the form read/write/delete:name" collectionFormat(multi)
// @Success      200  {object}  api.MessageResponse
// @Failure      400  {object}  api.BadRequestResponse
// @Failure      500  {object}  api.ServerErrorResponse
// @Router       /users/permissions/{id} [delete]
func (ue UserEndpoint) RevokePermissions(c *gin.Context) {
	userIDString := c.Param("id")
	id, err := api.GetUserIDFromString(userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	permissions, err := queryPermissionsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	err = ue.userRepo.RevokePermissions(id, permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "failed to revoke permissions: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "permissions revoked"})
}

// @Summary     List user permissions
// @Description Lists the permissions of a user by their ID.
// @Tags        users
// @Produce     application/json
// @Param       id path int true "User ID"
// @Success     200 {object} data.Permissions
// @Failure     400 {object} api.BadRequestResponse
// @Failure     500 {object} api.ServerErrorResponse
// @Router      /users/permissions/{id} [get]
func (ue UserEndpoint) GetPermissions(c *gin.Context) {
	id, err := api.GetUserIDFromString(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	perms, err := ue.userRepo.GetUserPermissions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

func queryPermissionsFromRequest(c *gin.Context) (data.Permissions, error) {
	permissions := c.QueryArray("permission")
	if len(permissions) == 0 {
		return nil, fmt.Errorf("no permissions provided")
	}
	var perms data.Permissions
	for _, perm := range permissions {
		permission, ok := data.GetPermissionFromSlug(perm)
		if !ok {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
		perms = append(perms, permission)
	}
	return perms, nil
}
