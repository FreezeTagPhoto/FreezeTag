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

// @summary     Add image to album, user must have write access to the album
// @description Add an image to an album
// @tags        albums
// @router      /album/add_image [post]
// @param request body AlbumImageRequest true "Add Image to Album Payload"
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
	var req AlbumImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}
	err = ae.albumRepository.SetImageAlbum(database.ImageId(req.ImageId), database.AlbumID(req.AlbumId), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Image added to album successfully"})
}

// @summary     Remove image from album, user must have write access to the album
// @description Remove an image from an album
// @tags        albums
// @router      /album/remove_image [post]
// @param request body AlbumImageRequest true "Remove Image from Album Payload"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) RemoveImageFromAlbum(c *gin.Context) {
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
	var req AlbumImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}
	err = ae.albumRepository.RemoveImageFromAlbum(database.ImageId(req.ImageId), database.AlbumID(req.AlbumId), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Image removed from album successfully"})
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
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Album visibility updated successfully"})
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

// @summary     List visible albums with ids
// @description Lists visible albums with id, name, owner and explicit shared users (owner only).
// @tags        albums
// @router      /album/list [get]
// @success     200 {array} database.Album
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) ListAlbums(c *gin.Context) {
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
	albums, err := ae.albumRepository.GetAlbums(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

func (ae AlbumEndpoint) ListAlbumImages(c *gin.Context) {
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

	albumID, err := api.ParseParamIntoID[database.AlbumID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	imageIDs, err := ae.albumRepository.GetAlbumImages(albumID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, imageIDs)
}

// @summary    	List albums associated with image
// @description List albums associated with an image, including id, name and owner of each album.
// @tags       	albums
// @router     	/album/list_by_image/{id} [get]
// @param id path int true "Image ID"
// @success    	200 {array} database.Album
// @failure    	400 {object} api.BadRequestResponse
// @failure    	500 {object} api.ServerErrorResponse
// @produce    	application/json
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

	albums, err := ae.albumRepository.GetAssociatedAlbums(imageID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

// @summary     Rename album
// @description Rename an album by providing the old name and new name. User must be an owner of the album.
// @tags        albums
// @router      /album/rename [post]
// @param request body RenameAlbumRequest true "Rename Album Payload"
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) RenameAlbum(c *gin.Context) {
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

	var req RenameAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: "invalid request body"})
		return
	}

	err = ae.albumRepository.RenameAlbum(req.AlbumID, req.NewName, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Album renamed successfully"})
}

// @summary     Delete album
// @description Delete an album by providing the album name. User must be an owner of the album.
// @tags        albums
// @router      /album/delete [delete]
// @success     200 {object} api.MessageResponse
// @failure     400 {object} api.BadRequestResponse
// @failure     500 {object} api.ServerErrorResponse
// @produce     application/json
func (ae AlbumEndpoint) DeleteAlbum(c *gin.Context) {
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

	albumId := c.Param("id")
	albumID, err := api.ParseParamIntoID[database.AlbumID](albumId)

	err = ae.albumRepository.RemoveAlbum(albumID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Album deleted successfully"})
}

func (ae AlbumEndpoint) GetAlbumInfo(c *gin.Context) {
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

	albumID, err := api.ParseParamIntoID[database.AlbumID](c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.BadRequestResponse{Error: err.Error()})
		return
	}

	album, err := ae.albumRepository.GetAlbum(albumID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ServerErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, album)
}
