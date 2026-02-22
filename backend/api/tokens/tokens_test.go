package tokens

import (
	"encoding/json"
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockService "freezetag/backend/mocks/AuthService"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func (te *TokenEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("tokens/revoke/:id", te.RevokeUserToken)
	e.POST("tokens/admin/revoke/:id", te.AdminRevokeToken)
	e.DELETE("tokens/delete/:id", te.AdminDeleteUserToken)
	e.POST("tokens/create", te.CreateUserToken)
}

func TestRevokeUserTokenInvalidTokenID(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)


	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/revoke/sdfhb", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestRevokeUserTokenUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestRevokeUserTokenSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().RevokeAPIToken(database.UserID(1), database.TokenID(1)).Return(nil).Once()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.MessageResponse{Message: "Token revoked successfully"}
	assert.Equal(t, expected, got)
}

func TestRevokeUserTokenDatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().RevokeAPIToken(database.UserID(1), database.TokenID(1)).Return(assert.AnError).Once()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestAdminDeleteUserTokenInvalidTokenID(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("DELETE", "/tokens/delete/sdfhb", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestRevokeUserTokenBadUserId(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "abc")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestAdminRevokeToken(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().AdminRevokeAPIToken(database.TokenID(1)).Return(nil).Once()
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/admin/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.MessageResponse{Message: "Token revoked successfully"}
	assert.Equal(t, expected, got)
}

func TestAdminRevokeTokenInvalidTokenID(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/admin/revoke/sdfhb", nil)
	router.ServeHTTP(w, req)
	
}

func TestCreateTokenSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.
		EXPECT().
		CreateAPIToken(database.UserID(1), data.Permissions{data.ReadFiles}, (*time.Time)(nil), "test token").
		Return(
			services.ApiCreateToken{
				TokenId: 1, TokenString: "plaintexttoken",
		}, nil).
		Once()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/create?permission=read:files&label=test token", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got services.ApiCreateToken
	err := json.Unmarshal(w.Body.Bytes(), &got)
	t.Log(w.Body.String())
	assert.NoError(t, err)
	expected := services.ApiCreateToken{TokenId: 1, TokenString: "plaintexttoken"}
	assert.Equal(t, expected, got)
}

func TestCreateUserTokenNoUserId(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()

	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/create?permission=read:files&label=test token", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCreateUserTokenInvalidPermission(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/create?permission=invalid:perm&label=test token", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCreateUserTokenInvalidUserId(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "abc")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/create?permission=read:files&label=test token", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCreateUserTokenShouldBindQueryError(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/create?permission=read:files&expiresAt=invalid&label=test token", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.BadRequestResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCreateUserTokenDatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.
		EXPECT().
		CreateAPIToken(database.UserID(1), data.Permissions{data.ReadFiles}, (*time.Time)(nil), "test token").
		Return(services.ApiCreateToken{}, assert.AnError).
		Once()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "1")
		c.Next()
	})
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/create?permission=read:files&label=test token", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestAdminDeleteUserToken(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().DeleteAPIToken(database.TokenID(1)).Return(nil).Once()
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("DELETE", "/tokens/delete/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.MessageResponse{Message: "Token deleted successfully"}
	assert.Equal(t, expected, got)
}

func TestAdminRevokeUserToken(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().AdminRevokeAPIToken(database.TokenID(1)).Return(nil).Once()
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/admin/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got api.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.MessageResponse{Message: "Token revoked successfully"}
	assert.Equal(t, expected, got)
}

func TestAdminRevokeUserTokenDatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().AdminRevokeAPIToken(database.TokenID(1)).Return(assert.AnError).Once()
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("POST", "/tokens/admin/revoke/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}	


func TestAdminDeleteUserTokenDatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	authService := mockService.NewMockAuthService(t)
	authService.EXPECT().DeleteAPIToken(database.TokenID(1)).Return(assert.AnError).Once()
	router := gin.New()
	te := InitTokenEndpoint(authService)
	te.RegisterEndpoints(router)

	req, _ := http.NewRequest("DELETE", "/tokens/delete/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got api.ServerErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}	

