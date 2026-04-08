package user

import (
	"fmt"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserEndpoint struct {
	authService services.AuthService
}

func InitUserEndpoint(authService services.AuthService) UserEndpoint {
	return UserEndpoint{authService: authService}
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
	userID, err := api.ParseParamIntoID[database.UserID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "Invalid user ID parameter"})
		return
	}
	user, err := ue.authService.GetUserByID(userID)
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
	users, err := ue.authService.AllUsers()
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
	id, err := api.ParseParamIntoID[database.UserID](userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	err = ue.authService.DeleteUser(id)
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
	id, err := api.ParseParamIntoID[database.UserID](userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	permissions, err := api.QueryPermissionsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	err = ue.authService.GrantPermissions(id, permissions)
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
	id, err := api.ParseParamIntoID[database.UserID](userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	permissions, err := api.QueryPermissionsFromRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	err = ue.authService.RevokePermissions(id, permissions)
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
	id, err := api.ParseParamIntoID[database.UserID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	perms, err := ue.authService.GetUserPermissions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// @Summary     Get a user's profile picture
// @Description Retrieves the profile picture of a user by their ID.
// @Description Retrieves the profile picture of a user by their ID.
// a user can only access their own profile picture, but an admin with the appropriate permissions can access any user's profile picture.
// @Tags        users
// @Produce     image/webp
// @Param       id path int true "User ID"
// @Success     200 {file} string "profile picture file data"
// @Failure     400 {object} api.BadRequestResponse
// @Failure     500 {object} api.ServerErrorResponse
// @Router      /users/profile-picture/{id} [get]
func (ue UserEndpoint) GetProfilePicture(c *gin.Context) {
	id, err := api.ParseParamIntoID[database.UserID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}
	picture, err := ue.authService.GetUserProfilePicture(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.Data(http.StatusOK, "image/webp", picture)
}

// @Summary     Update a user's profile picture
// @Description Updates the profile picture of a user by their ID. Accepts a multipart form with a "picture" file field. a user can only update their own profile picture,
// but an admin with the appropriate permissions can update any user's profile picture. The picture will be converted to webp format if it is not already in that format.
// @Tags        users
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       id   path      int  true  "User ID"
// @Param       picture formData file true "New profile picture"
// @Success     200 {object} api.MessageResponse
// @Failure     400 {object} api.BadRequestResponse
// @Failure     500 {object} api.ServerErrorResponse
// @Router      /users/profile-picture/{id} [post]
func (ue UserEndpoint) SetProfilePicture(c *gin.Context) {
	targetID, err := api.ParseParamIntoID[database.UserID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid target id"})
		return
	}

	file, err := c.FormFile("picture")
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to get picture from form data: " + err.Error()})
		return
	}
	bytes, err := api.ReadFileBytes(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "failed to read picture bytes: " + err.Error()})
		return
	}
	if err := ue.authService.SetUserProfilePicture(targetID, bytes, file.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "failed to update profile picture: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "profile picture updated"})
}

// @Summary     Set a user's visibility mode
// @Description Sets the visibility mode of a user by their ID. The visibility mode determines who can see the user's profile and images. Accepts a query parameter "mode" with values 0, 1, or 2, where 0 means only the user can see their profile and images, 1 means anyone can see their profile but only followers can see their images, and 2 means anyone can see their profile and images. a user can only update their own visibility mode, but an admin with the appropriate permissions can update any user's visibility mode.
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Param       id   path      int  true  "User ID"
// @Param       mode query     int  true  "Visibility mode (0, 1, or 2)"
// @Success     200 {object} api.MessageResponse
// @Failure     400 {object} api.BadRequestResponse
// @Failure     500 {object} api.ServerErrorResponse
// @Router      /users/visibility/{id} [post]
func (ue UserEndpoint) SetUserVisibilityMode(c *gin.Context) {
	targetID, err := api.ParseParamIntoID[database.UserID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid target id"})
		return
	}
	modeParam := c.Query("mode")
	if modeParam == "" {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "missing mode query parameter"})
		return
	}

	mode, err := strconv.Atoi(modeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid mode query parameter"})
		return
	}
	if mode < 0 || mode > 2 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "mode query parameter must be 0, 1, or 2"})
		return
	}
	if err := ue.authService.SetUserVisibilityMode(targetID, database.UserPrivacy(mode)); err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "failed to update user visibility mode: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "user visibility mode updated"})
}

// @Summary     Reset a user's profile picture to default
// @Description Resets the profile picture of a user to the default generated avatar. A user can only reset their own profile picture,
// but an admin with the appropriate permissions can reset any user's profile picture.
// @Tags        users
// @Produce     application/json
// @Param       id path int true "User ID"
// @Success     200 {object} api.MessageResponse
// @Failure     400 {object} api.BadRequestResponse
// @Failure     500 {object} api.ServerErrorResponse
// @Router      /users/profile-picture/{id} [delete]
func (ue UserEndpoint) ResetProfilePicture(c *gin.Context) {
	targetID, err := api.ParseParamIntoID[database.UserID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid target id"})
		return
	}
	if err := ue.authService.ResetProfilePicture(targetID); err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "failed to reset profile picture: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "profile picture reset to default"})
}
