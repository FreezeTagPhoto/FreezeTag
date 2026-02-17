package permissions

import (
	"encoding/json"
	"freezetag/backend/pkg/database/data"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPermissions(t *testing.T) {
	w := httptest.NewRecorder()
	router := gin.Default()
	pe := initPermissionEndpoint()
	router.GET("/permissions/list", pe.ListPermissions)

	req, _ := http.NewRequest("GET", "/permissions/list", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)	
	var gotPermissions []data.Permission
	err := json.Unmarshal(w.Body.Bytes(), &gotPermissions)
	require.NoError(t, err)

	expectedPermissions := data.AllPermissions()
	assert.ElementsMatch(t, expectedPermissions, gotPermissions)
}