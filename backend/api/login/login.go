package login

import (
	"freezetag/backend/api"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoginEndpoint struct {
	authService services.AuthService
}

func InitLoginEndpoint(authService services.AuthService) LoginEndpoint {
	return LoginEndpoint{
		authService: authService,
	}
}



// @summary Authenticate user and return a token
// @description Authenticates a user using form parameters "username" and "password". On success returns a JSON payload containing an authentication token.
// @tags auth, login
// @accept application/json
// @produce application/json
// @param request body api.LoginCredentials true "User Login Details"
// @success 200 {object} api.StatusLoginSuccess "Authentication successful"
// @failure 401 {object} api.StatusLoginFail "Authentication failed"
// @router /login [post]
func (le LoginEndpoint) HandleLogin(c *gin.Context) {
	var req api.LoginCredentials
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.StatusBadRequestResponse{Error: "invalid request: " + err.Error()})
		return
	}

	token, err := le.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.StatusLoginFail{Error: "authentication failed: " + err.Error()})
		return
	}
	c.SetCookieData(&http.Cookie{
		Name:     "token",
		Value:    token,
		Secure:   false,
		HttpOnly: true,
	})
	c.JSON(http.StatusOK, api.StatusLoginSuccess{Token: token})
}

// @summary Check login status
// @description Checks if the user is currently authenticated.
// @tags auth, login
// @produce application/json
// @success 200 {object} api.StatusLoginUser "user is authenticated"
// @failure 401 {object} api.StatusLoginFail "User is not authenticated"
// @router /login [get]
func (le LoginEndpoint) HandleLoginStatus(c *gin.Context) {
	authenticated, err := c.Cookie("token")
	if err != nil || authenticated == "" {
		c.JSON(http.StatusUnauthorized, api.StatusLoginFail{Error: "not authenticated"})
		return
	}
	claims, err := le.authService.ValidateJWT(authenticated)
	if err != nil {
		c.JSON(http.StatusUnauthorized, api.StatusLoginFail{Error: "not authenticated"})
		return
	}


	if claims.Subject == "" {
		// Handle the case where "sub" isn't a valid number
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
		return
	}

	uid, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
		return
	}

	id := database.UserID(uid)
	c.JSON(http.StatusOK, api.StatusLoginUser{UserID: id})
}
