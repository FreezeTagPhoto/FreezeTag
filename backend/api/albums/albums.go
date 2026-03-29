package albums

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlbumEndpoint struct {
	albumRepository database.AlbumDatabase
}

type AlbumCreateRequest struct {
	Name           string                `json:"name" binding:"required"`
	VisibilityMode database.PrivacyLevel `json:"visibility_mode" binding:"oneof=0 1 2"`
}

type AlbumModifyRequest struct {
	AlbumId        database.AlbumId       `json:"album_id" binding:"required"`
	Name           *string                `json:"name,omitempty"`
	VisibilityMode *database.PrivacyLevel `json:"visibility_mode,omitempty" binding:"omitempty,oneof=0 1"`
}

type UserAlbumPermissionRequest struct {
	AlbumId      database.AlbumId      `json:"album_id" binding:"required"`
	TargetUserId database.UserID       `json:"target_user_id" binding:"required"`
	Permission   database.PrivacyLevel `json:"permission" binding:"oneof=0 1 2"`
}

type AddImageRequest struct {
	ImageId database.ImageId `json:"image_id"`
	AlbumId database.AlbumId `json:"album_id"`
}

type AddImageByNameRequest struct {
	ImageId   database.ImageId `json:"image_id"`
	AlbumName string           `json:"album_name" binding:"required"`
}

func InitAlbumEndpoint(repository database.AlbumDatabase) AlbumEndpoint {
	return AlbumEndpoint{
		albumRepository: repository,
	}
}

// @summary     Create album
// @description Create a new album with a given name and visibility mode
// @tags        albums
// @router      /album/create [post]
// @param request body AlbumCreateRequest true "Create Album Payload"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) CreateAlbum(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}
	var req AlbumCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}

	// Call the repository method to create the album
	albumID, err := ae.albumRepository.CreateAlbum(req.Name, uid, req.VisibilityMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.AlbumCreateResponse{AlbumID: albumID})
}

// @summary     Add image to album
// @description Add an image to an album
// @tags        albums
// @router      /album/add_image [post]
// @param request body AddImageRequest true "Add Image to Album Payload"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) AddImageToAlbum(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}
	var req AddImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}
	err = ae.albumRepository.SetImageAlbum(database.ImageId(req.ImageId), database.AlbumId(req.AlbumId), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Image added to album successfully"})
}

func (ae AlbumEndpoint) AddImageToAlbumByName(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}

	var req AddImageByNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}

	albumID, err := ae.albumRepository.GetAlbumIdByName(req.AlbumName, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	if albumID < 0 {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "album not found"})
		return
	}

	err = ae.albumRepository.SetImageAlbum(req.ImageId, albumID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Image added to album successfully"})
}

// @summary     Set album visibility
// @description Set the visibility mode of an album
// @tags        albums
// @router      /album/set_visibility [post]
// @param request body AlbumModifyRequest true "Set Album Visibility Payload"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) SetAlbumVisibility(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}
	var req AlbumModifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}
	if req.VisibilityMode == nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "visibility_mode is required"})
		return
	}
	err = ae.albumRepository.SetAlbumVisibility(req.AlbumId, *req.VisibilityMode, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
}

// @summary     Set user album permission
// @description Set the permission level of a user for a specific album
// @tags        albums
// @router      /album/set_permission [post]
// @param request body UserAlbumPermissionRequest true "Set User Album Permission Payload"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) SetUserAlbumPermission(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}
	var req UserAlbumPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}
	err = ae.albumRepository.SetUserAlbumPermission(req.AlbumId, req.TargetUserId, req.Permission, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "User album permission updated successfully"})
}

func (ae AlbumEndpoint) ListVisibleAlbums(c *gin.Context) {
	id, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}
	albums, err := ae.albumRepository.GetAlbumNames(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

func (ae AlbumEndpoint) ListImageAlbums(c *gin.Context) {
	id, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID not found in context"})
		return
	}
	uid, err := api.ParseParamIntoID[database.UserID](id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: "userID in context is not of type UserID"})
		return
	}

	imageID, err := api.ParseParamIntoID[database.ImageId](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	albums, err := ae.albumRepository.GetImageAlbumNames(imageID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

// @summary     Get album images
// @description Retrieve a list of image IDs contained in a specific album
// @tags        albums
// @router      /album/images [get]
// @param album_id query int true "Album ID"
// @success     200 {array} database.ImageId
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) GetAlbumImages(c *gin.Context) {

}
