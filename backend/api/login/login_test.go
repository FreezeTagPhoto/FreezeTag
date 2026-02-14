package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"freezetag/backend/api"
	mockUserService "freezetag/backend/mocks/AuthService"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type badLoginCredentials struct {
	Wrongtype string
}

func (le LoginEndpoint) RegisterEndpoints(e gin.IRoutes) {
	e.POST("/login", le.Login)
	e.GET("/login", le.LoginInfo)
}

func TestLogin(t *testing.T) {
	plaintextPassword := "securepassword"

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		AuthenticateUser("testuser", plaintextPassword).
		Return("json_token", nil).Once()

	router := gin.Default()
	jsonBytes, err := json.Marshal(api.LoginCredentials{
		Username: "testuser",
		Password: plaintextPassword,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/login", bytes.NewReader(jsonBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	got := w.Body.Bytes()
	var gotResponse api.StatusLoginSuccess
	err = json.Unmarshal(got, &gotResponse)
	assert.NoError(t, err)
	expected := api.StatusLoginSuccess{Token: "json_token"}
	assert.Equal(t, expected, gotResponse)

	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.Equal(t, "json_token", cookies[0].Value)
}

func TestLoginFailure(t *testing.T) {
	plaintextPassword := "securepassword"

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		AuthenticateUser("testuser", plaintextPassword).
		Return("", fmt.Errorf("authentication failed")).Once()

	router := gin.Default()
	loginCredentials := api.LoginCredentials{
		Username: "testuser",
		Password: plaintextPassword,
	}
	jsonBytes, err := json.Marshal(loginCredentials)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/login", bytes.NewReader(jsonBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var got api.StatusLoginFail
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.StatusLoginFail{Error: "authentication failed: authentication failed"}
	assert.Equal(t, expected, got)
}

func TestLoginBadCredentialFormat(t *testing.T) {
	NewMockAuthService := mockUserService.NewMockAuthService(t)
	router := gin.Default()
	loginCredentials := badLoginCredentials{
		Wrongtype: "bad",
	}

	jsonBytes, err := json.Marshal(loginCredentials)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/login", bytes.NewReader(jsonBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got api.StatusBadRequestResponse
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Contains(t, got.Error, "invalid request")
}

func TestLoginExistingLoginCookie(t *testing.T) {

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		ValidateJWT("existing_token").
		Return(services.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "1",
			},
		}, nil).Once()

	router := gin.Default()

	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: "existing_token",
	})
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.StatusLoginUser
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.StatusLoginUser{UserID: 1}
	assert.Equal(t, expected, got)
}

func TestLoginExistingLoginCookieNoAuth(t *testing.T) {

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		ValidateJWT("existing_token").
		Return(services.Claims{}, errors.New("no auth")).Once()

	router := gin.Default()

	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: "existing_token",
	})
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var got api.StatusLoginFail
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.StatusLoginFail{Error: "not authenticated"}
	assert.Equal(t, expected, got)
}

func TestLoginExistingLoginCookieNoSub(t *testing.T) {

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		ValidateJWT("existing_token").
		Return(services.Claims{}, nil).Once()

	router := gin.Default()

	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: "existing_token",
	})
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var got api.StatusLoginFail
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.StatusLoginFail{Error: "invalid user ID in token"}
	assert.Equal(t, expected, got)
}

func TestLoginNoExistingLoginCookie(t *testing.T) {

	NewMockAuthService := mockUserService.NewMockAuthService(t)

	router := gin.Default()

	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var got api.StatusLoginFail
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.StatusLoginFail{Error: "not authenticated"}
	assert.Equal(t, expected, got)
}

func TestLoginInfoBadId(t *testing.T) {
	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		ValidateJWT("existing_token").
		Return(services.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "1",
			},
			Permissions: data.Permissions{"permission"},
		}, nil).Once()

	router := gin.Default()

	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "token",
		Value: "existing_token",
	})
	InitLoginEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got api.StatusLoginUser
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	expected := api.StatusLoginUser{UserID: 1, Permissions: data.Permissions{"permission"}}
	assert.Equal(t, expected, got)
}
