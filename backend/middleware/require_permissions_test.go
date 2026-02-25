package middleware

import (
	"freezetag/backend/pkg/database/data"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPermissionsMiddleware(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", data.Permissions{data.ReadFiles, data.WriteFiles, data.WriteUser})
	RequirePermission(data.ReadFiles, data.WriteFiles)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware, but it was aborted")
	}
}

func TestPermissionsMiddlewareInvalidPermissions(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", data.Permissions{})
	RequirePermission(data.ReadFiles, data.WriteFiles)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to be aborted by middleware, but it passed through")
	}
}

func TestPermissionsMiddlewareInvalidPermissionsType(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", 1)
	RequirePermission(data.ReadFiles, data.WriteFiles)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to be aborted by middleware due to invalid permissions type, but it passed through")
	}
}
func TestPermissionsMiddlewareNoPermissions(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	RequirePermission(data.ReadFiles, data.WriteFiles)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to be aborted by middleware due to invalid permissions type, but it passed through")
	}
}

func TestPermissionsMiddlewareOrSelf(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", data.Permissions{})
	ctx.Set("userID", "123")
	ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: "123"})
	RequirePermissionOrSelf(data.ReadFiles, data.WriteFiles)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware due to self access, but it was aborted")
	}
}

func TestPermissionsMiddlewareOrSelfNoPermissions(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", data.Permissions{})
	ctx.Set("userID", "123")
	ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: "123"})
	RequirePermissionOrSelf(data.ReadFiles, data.WriteFiles)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware due to self access, but it was aborted")
	}
}

func TestPermissionsMiddlewareOrSelfInvalidPermissionsType(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", 1)
	ctx.Set("userID", "123")
	ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: "123"})
	RequirePermissionOrSelf(data.ReadFiles, data.WriteFiles)(ctx)
	if ctx.IsAborted() {
		t.Errorf("Expected request to pass through middleware due to self access, but it was aborted")
	}
}

func TestPermissionsMiddlewareOrSelfNoUserID(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", data.Permissions{})
	RequirePermissionOrSelf(data.ReadFiles, data.WriteFiles)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to be aborted by middleware due to no user ID and insufficient permissions, but it passed through")
	}
}

func TestPermissionsMiddlewareOrSelfInvalidUserIDType(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/", nil)
	ctx.Request = req
	ctx.Set("permissions", data.Permissions{})
	ctx.Set("userID", 123)
	RequirePermissionOrSelf(data.ReadFiles, data.WriteFiles)(ctx)
	if !ctx.IsAborted() {
		t.Errorf("Expected request to be aborted by middleware due to invalid user ID type and insufficient permissions, but it passed through")
	}
}
