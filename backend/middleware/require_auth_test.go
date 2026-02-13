package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockUserService "freezetag/backend/mocks/AuthService"
	"freezetag/backend/pkg/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
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
	require.True(t, exists)
	require.Equal(t, "expectedUserID", userID)
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
	require.False(t, exists)
	require.Equal(t, nil, userID)
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
	require.False(t, exists)
	require.Equal(t, nil, userID)
}
