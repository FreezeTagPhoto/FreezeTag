package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockUserService "freezetag/backend/mocks/AuthService"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "TOKENSTRING"})
	ctx.Request = req

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	claims := services.Claims{
		Permissions: nil,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "expectedUserID", // Correctly nesting the embedded field
		},
	}
	NewMockAuthService.EXPECT().
		ValidateJWT("TOKENSTRING").
		Return(claims, nil).Once()

	RequireAuth(NewMockAuthService)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware, but it was aborted")
	}

	userID, exists := ctx.Get("userID")
	assert.True(t, exists)
	assert.Equal(t, "expectedUserID", userID)
}

func TestAuthMiddlewareJWTparseFail(t *testing.T) {

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "TOKENSTRING"})
	ctx.Request = req

	NewMockAuthService := mockUserService.NewMockAuthService(t)

	NewMockAuthService.EXPECT().
		ValidateJWT("TOKENSTRING").
		Return(services.Claims{}, errors.New("an error")).Once()

	RequireAuth(NewMockAuthService)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to abort middleware, but it was not")
	}

	userID, exists := ctx.Get("userID")
	assert.False(t, exists)
	assert.Equal(t, nil, userID)
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req

	NewMockAuthService := mockUserService.NewMockAuthService(t)

	RequireAuth(NewMockAuthService)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to abort middleware, but it was not")
	}

	userID, exists := ctx.Get("userID")
	assert.False(t, exists)
	assert.Equal(t, nil, userID)
}

func TestAuthMiddlewareNoSub(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "TOKENSTRING"})
	ctx.Request = req

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	claims := services.Claims{
		Permissions: nil,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "", // No subject
		},
	}
	NewMockAuthService.EXPECT().
		ValidateJWT("TOKENSTRING").
		Return(claims, nil).Once()

	RequireAuth(NewMockAuthService)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to abort middleware, but it was not")
	}

	userID, exists := ctx.Get("userID")
	assert.False(t, exists)
	assert.Equal(t, nil, userID)
}

func TestAuthMiddlewareHasPermissions(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "TOKENSTRING"})
	ctx.Request = req

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	claims := services.Claims{
		Permissions: data.Permissions{data.CreateUser, data.ReadUser},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "expectedUserID",
		},
	}
	NewMockAuthService.EXPECT().
		ValidateJWT("TOKENSTRING").
		Return(claims, nil).Once()

	RequireAuth(NewMockAuthService)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware, but it was aborted")
	}
	permissions, exists := ctx.Get("permissions")
	assert.True(t, exists)
	assert.ElementsMatch(t, data.Permissions{data.CreateUser, data.ReadUser}, permissions)
}

func TestAuthMiddlewareHasNoPermissions(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "TOKENSTRING"})
	ctx.Request = req

	NewMockAuthService := mockUserService.NewMockAuthService(t)
	claims := services.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "expectedUserID",
		},
	}
	NewMockAuthService.EXPECT().
		ValidateJWT("TOKENSTRING").
		Return(claims, nil).Once()

	RequireAuth(NewMockAuthService)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware, but it was aborted")
	}
	permissions, exists := ctx.Get("permissions")
	assert.True(t, exists)
	// we dont want permissions to be nil otherwise its a pain
	assert.ElementsMatch(t, data.Permissions{}, permissions)
}
