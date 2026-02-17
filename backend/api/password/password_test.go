package password

import (
	"bytes"
	"encoding/json"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"net/http"
	"net/http/httptest"
	"testing"

	mockUserService "freezetag/backend/mocks/AuthService"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (pe PasswordEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("/password/change", pe.ChangePassword)
}

func TestChangePassword(t *testing.T) {
	w := httptest.NewRecorder()
	mockAuthService := mockUserService.NewMockAuthService(t)
	mockAuthService.EXPECT().ChangePassword(database.UserID(1), "oldpassword", "newpassword").Return(nil).Once()
	router := gin.New()

	// simulate middleware here so that userID gets set correctly
	// could also just call the handler function directly with a proper context, but this is closer to
	// a real request
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	InitPasswordEndpoint(mockAuthService).RegisterEndpoints(router)
	reqBody := api.PasswordChangeRequest{
		Username:        "testuser",
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword",
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)
	req, err := http.NewRequest("POST", "/password/change", bytes.NewBuffer(reqBodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got api.MessageResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)

	expected := api.MessageResponse{Message: "password changed successfully"}
	assert.Equal(t, expected, got)
}

func TestChangePasswordInvalidRequest(t *testing.T) {
	w := httptest.NewRecorder()
	mockAuthService := mockUserService.NewMockAuthService(t)
	router := gin.New()
	InitPasswordEndpoint(mockAuthService).RegisterEndpoints(router)
	req, err := http.NewRequest("POST", "/password/change", bytes.NewBuffer([]byte("invalid json")))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got api.BadRequestResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)

	assert.Contains(t, got.Error, "invalid request")
}
