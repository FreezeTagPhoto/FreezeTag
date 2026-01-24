package createuser

import (
	"bytes"
	"encoding/json"
	"freezetag/backend/api"
	mockUserService "freezetag/backend/mocks/AuthService"
	"freezetag/backend/pkg/database"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	plaintextPassword := "securepassword"
	expectedUser := &database.PublicUser{
		ID:        1,
		Username:  "newuser",
		CreatedAt: time.Now().Unix(),
	}

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	NewMockAuthService.EXPECT().
		AddUser("newuser", plaintextPassword).
		Return(expectedUser, nil).Once()

	router := gin.Default()
	loginCredentials := api.LoginCredentials{
		Username: "newuser",
		Password: plaintextPassword,
	}
	jsonBytes, err := json.Marshal(loginCredentials)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/createuser", bytes.NewReader(jsonBytes))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	InitCreateUserEndpoint(NewMockAuthService).RegisterEndpoints(router)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got database.PublicUser
	err = json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, *expectedUser, got)
}
