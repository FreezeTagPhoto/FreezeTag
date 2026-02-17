package logout

import (
	"encoding/json"
	"freezetag/backend/api"
	mockUserService "freezetag/backend/mocks/AuthService"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (lo LogoutEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("/logout", lo.HandleLogout)
}

func TestLogout(t *testing.T) {

	router := gin.Default()
	w := httptest.NewRecorder()

	NewMockAuthService := mockUserService.NewMockAuthService(t)

	req, err := http.NewRequest("POST", "/logout", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "token", Value: "TOKENSTRING"})
	InitLogoutEndpoint(NewMockAuthService).RegisterEndpoints(router)

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, -1, w.Result().Cookies()[0].MaxAge)
	assert.Equal(t, "", w.Result().Cookies()[0].Value)
	var got api.LogoutSuccessResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.LogoutSuccessResponse{Status: "ok"}
	assert.Equal(t, expected, got)
}
