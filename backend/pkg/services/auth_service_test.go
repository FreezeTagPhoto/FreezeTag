package services

import (
	"fmt"
	mockUserDatabase "freezetag/backend/mocks/UserDatabase"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAddUser(t *testing.T) {
	plaintextPassword := "securepassword"
	dummyHash := "dummyHash"

	user := &database.PublicUser{
		ID:           database.UserID(42),
		Username:     "newuser",
		CreatedAt:    time.Now().Unix(),
		PasswordHash: dummyHash,
	}

	// bcrypt automatically salts the password, so we can't predict the hash value.
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		AddUser("newuser", mock.AnythingOfType("string")).
		RunAndReturn(
			func(username string, passwordHash string) (*database.PublicUser, error) {
				user.PasswordHash = passwordHash
				return user, nil
			}).
		Once()

	authService := InitDefaultAuthService(mockDb)
	userGot, err := authService.AddUser("newuser", plaintextPassword)
	require.NoError(t, err)
	require.Equal(t, user, userGot)
	assert.NotEqual(t, dummyHash, userGot.PasswordHash)

	err = bcrypt.CompareHashAndPassword([]byte(userGot.PasswordHash), []byte(plaintextPassword))
	require.NoError(t, err, "Password hash does not match the original password")
}

func TestEnsureLoginNoUsers(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().AllUsers().Return(nil, nil)
	mockDb.EXPECT().AddUser("admin", mock.AnythingOfType("string")).Return(&database.PublicUser{ID: 1}, nil)
	mockDb.EXPECT().EnsureAdmin(mock.Anything).Return(nil)
	authService := InitDefaultAuthService(mockDb)
	err := authService.EnsureLogin()
	assert.NoError(t, err)
}

func TestEnsureLoginAlreadyUser(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().AllUsers().Return([]database.PublicUser{{ID: 1}}, nil)
	authService := InitDefaultAuthService(mockDb)
	err := authService.EnsureLogin()
	assert.NoError(t, err)
}

func TestAddUserFails(t *testing.T) {
	plaintextPassword := "securepassword"

	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		AddUser("domer", mock.AnythingOfType("string")).
		Return(nil, assert.AnError).
		Once()

	authService := InitDefaultAuthService(mockDb)
	userGot, err := authService.AddUser("domer", plaintextPassword)
	require.Error(t, err)
	assert.Nil(t, userGot)
}

func TestAuthenticateUser(t *testing.T) {
	plaintextPassword := "securepassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetUserByUsername("authuser").
		Return(&database.PublicUser{
			ID:           database.UserID(7),
			Username:     "authuser",
			CreatedAt:    time.Now().Unix(),
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()
	mockDb.EXPECT().
		GetUserPermissions(database.UserID(7)).
		Return(data.Permissions{}, nil).
		Once()

	authService := InitDefaultAuthService(mockDb)
	_, err = authService.AuthenticateUser("authuser", plaintextPassword)
	require.NoError(t, err)
}

func TestAuthenticateUserFails(t *testing.T) {
	plaintextPassword := "securepassword"
	wrongPassword := "wrongpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetUserByUsername("authuser").
		Return(&database.PublicUser{
			ID:           database.UserID(7),
			Username:     "authuser",
			CreatedAt:    time.Now().Unix(),
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()

	authService := InitDefaultAuthService(mockDb)
	_, err = authService.AuthenticateUser("authuser", wrongPassword)
	require.Error(t, err)
}

func TestAuthenticateNonexistentUser(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetUserByUsername("nonexistent").
		Return(nil, assert.AnError).
		Once()

	authService := InitDefaultAuthService(mockDb)
	_, err := authService.AuthenticateUser("nonexistent", "anyPassword")
	require.Error(t, err)
}

func TestCreateToken(t *testing.T) {
	userID := database.UserID(123)
	tokenString, err := createTokenWithPermissions(userID, data.Permissions{data.ReadUser})
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Parse the token to verify its contents
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	})
	require.NoError(t, err)
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, "123", claims["sub"])
}

func TestLoginCreatesValidJWT(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)

	uid := database.UserID(123)

	password := "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 1)
	mockDb.EXPECT().
		GetUserByUsername("testuser").
		Return(&database.PublicUser{
			ID:           uid,
			Username:     "testuser",
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()
	mockDb.EXPECT().
		GetUserPermissions(uid).
		Return(data.Permissions{data.ReadUser}, nil).
		Once()

	service := InitDefaultAuthService(mockDb)
	tokenString, err := service.AuthenticateUser("testuser", password)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	claims := JWTClaims{}
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	}, jwt.WithValidMethods([]string{JwtSigningMethod.Alg()}))
	require.NoError(t, err)

	t.Logf("%s", claims)
	sub := claims.Subject
	if fmt.Sprintf("%v", sub) != fmt.Sprintf("%d", uid) {
		t.Errorf("Expected sub claim %d, got %v", uid, sub)
	}
	JWTpermissions := claims.Permissions
	assert.True(t, JWTpermissions.HasPermission(data.ReadUser))
}

func TestValidateJWT(t *testing.T) {
	auth := InitDefaultAuthService(nil)
	userID := database.UserID(456)
	tokenString, err := createTokenWithPermissions(userID, data.Permissions{})
	require.NoError(t, err)
	claims, err := auth.ValidateJWT(tokenString)
	require.NoError(t, err)
	require.Equal(t, "456", claims.Subject)
}

func TestValidateJWTinvalidToken(t *testing.T) {
	auth := InitDefaultAuthService(nil)
	claims, err := auth.ValidateJWT("invalid token")
	require.Error(t, err)
	require.Equal(t, JWTClaims{}, claims)
}

func TestValidateJWTexpiredToken(t *testing.T) {
	auth := InitDefaultAuthService(nil)
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "789",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	})
	tokenString, err := expiredToken.SignedString([]byte(JwtSecretKey))
	require.NoError(t, err)
	claims, err := auth.ValidateJWT(tokenString)
	require.Error(t, err)
	require.Equal(t, JWTClaims{}, claims)
}

func TestCreateJWTNoPermissions(t *testing.T) {
	userID := database.UserID(123)
	tokenString, err := createTokenWithPermissions(userID, nil)
	require.NoError(t, err)
	claims := JWTClaims{}
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	}, jwt.WithValidMethods([]string{JwtSigningMethod.Alg()}))
	require.NoError(t, err)
	assert.Empty(t, claims.Permissions)
}

func TestEnsureLoginCreatesAdmin(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().AllUsers().Return(nil, nil)
	mockDb.EXPECT().AddUser("admin", mock.AnythingOfType("string")).Return(&database.PublicUser{ID: 1}, nil)
	mockDb.EXPECT().EnsureAdmin(database.UserID(1)).Return(nil)
	authService := InitDefaultAuthService(mockDb)
	err := authService.EnsureLogin()
	assert.NoError(t, err)
}

func TestEnsureLoginWithUsers(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().AllUsers().Return([]database.PublicUser{{ID: 1}}, nil)
	authService := InitDefaultAuthService(mockDb)
	err := authService.EnsureLogin()
	assert.NoError(t, err)
}

func TestEnsureLoginFailsAddUser(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().AllUsers().Return(nil, nil)
	mockDb.EXPECT().AddUser("admin", mock.AnythingOfType("string")).Return(nil, assert.AnError)
	authService := InitDefaultAuthService(mockDb)
	err := authService.EnsureLogin()
	assert.Error(t, err)
}

func TestEnsureLoginFailsGrantAdmin(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().AllUsers().Return(nil, nil)
	mockDb.EXPECT().AddUser("admin", mock.AnythingOfType("string")).Return(&database.PublicUser{ID: 1}, nil)
	mockDb.EXPECT().EnsureAdmin(database.UserID(1)).Return(assert.AnError)
	authService := InitDefaultAuthService(mockDb)
	err := authService.EnsureLogin()
	assert.Error(t, err)
}

func TestValidateAPIToken(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetApiUserID(mock.Anything).
		Return(database.UserID(1), nil).
		Once()
	mockDb.EXPECT().
		GetApiPermissions(mock.Anything).
		Return(data.Permissions{data.ReadUser}, nil).
		Once()
	authService := InitDefaultAuthService(mockDb)

	claims, err := authService.ValidateAPIToken("token")
	require.NoError(t, err)
	assert.ElementsMatch(t, data.Permissions{data.ReadUser}, claims.Permissions)
}

func TestValidateAPITokenInvalid(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetApiUserID(mock.Anything).
		Return(database.UserID(1), assert.AnError).
		Once()
	authService := InitDefaultAuthService(mockDb)

	claims, err := authService.ValidateAPIToken("invalid_token")
	require.Error(t, err)
	assert.Equal(t, ApiClaims{}, claims)
}

func TestHashToken(t *testing.T) {
	token := "sometoken"
	hashedToken := hashToken(token)
	require.NotEmpty(t, hashedToken)
	assert.NotEqual(t, token, hashedToken)
	hashedToken2 := hashToken(token)
	assert.Equal(t, hashedToken, hashedToken2)
}

func TestAuthenticateUserPermissionsError(t *testing.T) {
	plaintextPassword := "securepassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetUserByUsername("authuser").
		Return(&database.PublicUser{
			ID:           database.UserID(7),
			Username:     "authuser",
			CreatedAt:    time.Now().Unix(),
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()
	mockDb.EXPECT().
		GetUserPermissions(database.UserID(7)).
		Return(nil, assert.AnError).
		Once()

	authService := InitDefaultAuthService(mockDb)
	_, err = authService.AuthenticateUser("authuser", plaintextPassword)
	assert.Error(t, err)
}

func TestChangePasswordSuccess(t *testing.T) {
	plaintextPassword := "securepassword"
	newPassword := "newsecurepassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetPasswordHash(database.UserID(7)).
		Return(string(hashedPassword), nil).
		Once()
	mockDb.EXPECT().
		SetUserPassword(database.UserID(7), mock.AnythingOfType("string")).
		Return(nil).
		Run(func(userID database.UserID, newHash string) {
			err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte(newPassword))
			assert.NoError(t, err, "New password hash does not match the new password")
		}).
		Once()
	authService := InitDefaultAuthService(mockDb)
	err = authService.ChangePassword(database.UserID(7), plaintextPassword, newPassword)
	assert.NoError(t, err)
}

func TestChangePasswordInvalidHash(t *testing.T) {
	plaintextPassword := "securepassword"
	newPassword := "newsecurepassword"
	fakeHash, err := bcrypt.GenerateFromPassword([]byte("insecurePassword"), bcrypt.DefaultCost)
	require.NoError(t, err)
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetPasswordHash(database.UserID(7)).
		Return(string(fakeHash), nil).
		Once()
	authService := InitDefaultAuthService(mockDb)
	err = authService.ChangePassword(database.UserID(7), plaintextPassword, newPassword)
	assert.Error(t, err)
}

func TestUserRepoErrorChangePassword(t *testing.T) {
	plaintextPassword := "securepassword"
	newPassword := "newsecurepassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetPasswordHash(database.UserID(7)).
		Return(string(hashedPassword), nil).
		Once()
	mockDb.EXPECT().
		SetUserPassword(database.UserID(7), mock.AnythingOfType("string")).
		Return(assert.AnError).
		Once()
	authService := InitDefaultAuthService(mockDb)
	err = authService.ChangePassword(database.UserID(7), plaintextPassword, newPassword)
	assert.Error(t, err)
}

func TestForceChangePassword(t *testing.T) {
	newPassword := "newsecurepassword"
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		SetUserPassword(database.UserID(7), mock.AnythingOfType("string")).
		Return(nil).
		Run(func(userID database.UserID, newHash string) {
			err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte(newPassword))
			assert.NoError(t, err, "New password hash does not match the new password")
		}).
		Once()
	authService := InitDefaultAuthService(mockDb)
	err := authService.ForceChangePassword(database.UserID(7), newPassword)
	assert.NoError(t, err)
}

func TestForceChangePasswordRepoError(t *testing.T) {
	newPassword := "newsecurepassword"
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		SetUserPassword(database.UserID(7), mock.AnythingOfType("string")).
		Return(assert.AnError).
		Once()
	authService := InitDefaultAuthService(mockDb)
	err := authService.ForceChangePassword(database.UserID(7), newPassword)
	assert.Error(t, err)
}

func TestValidateAPITokenFailsGettingPermissions(t *testing.T) {
	mockDb := mockUserDatabase.NewMockUserDatabase(t)
	mockDb.EXPECT().
		GetApiUserID(mock.Anything).
		Return(database.UserID(1), nil).
		Once()
	mockDb.EXPECT().
		GetApiPermissions(mock.Anything).
		Return(nil, assert.AnError).
		Once()
	authService := InitDefaultAuthService(mockDb)

	claims, err := authService.ValidateAPIToken("token")
	require.Error(t, err)
	assert.Equal(t, ApiClaims{}, claims)
}
