package services

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"testing"
	"time"

	mockUserRepository "freezetag/backend/mocks/UserRepository"

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
	mockRepo := mockUserRepository.NewMockUserRepository(t)
	mockRepo.EXPECT().
		AddUser("newuser", mock.AnythingOfType("string")).
		RunAndReturn(
			func(username string, passwordHash string) (*database.PublicUser, error) {
				user.PasswordHash = passwordHash
				return user, nil
			}).
		Once()

	authService := InitDefaultAuthService(mockRepo)
	userGot, err := authService.AddUser("newuser", plaintextPassword)
	require.NoError(t, err)
	require.Equal(t, user, userGot)
	assert.NotEqual(t, dummyHash, userGot.PasswordHash)

	err = bcrypt.CompareHashAndPassword([]byte(userGot.PasswordHash), []byte(plaintextPassword))
	require.NoError(t, err, "Password hash does not match the original password")
}

func TestAddUserFails(t *testing.T) {
	plaintextPassword := "securepassword"

	mockRepo := mockUserRepository.NewMockUserRepository(t)
	mockRepo.EXPECT().
		AddUser("domer", mock.AnythingOfType("string")).
		Return(nil, assert.AnError).
		Once()

	authService := InitDefaultAuthService(mockRepo)
	userGot, err := authService.AddUser("domer", plaintextPassword)
	require.Error(t, err)
	assert.Nil(t, userGot)
}

func TestAuthenticateUser(t *testing.T) {
	plaintextPassword := "securepassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo := mockUserRepository.NewMockUserRepository(t)
	mockRepo.EXPECT().
		GetUserByUsername("authuser").
		Return(&database.PublicUser{
			ID:           database.UserID(7),
			Username:     "authuser",
			CreatedAt:    time.Now().Unix(),
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()

	authService := InitDefaultAuthService(mockRepo)
	_, err = authService.AuthenticateUser("authuser", plaintextPassword)
	require.NoError(t, err)
}

func TestAuthenticateUserFails(t *testing.T) {
	plaintextPassword := "securepassword"
	wrongPassword := "wrongpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo := mockUserRepository.NewMockUserRepository(t)
	mockRepo.EXPECT().
		GetUserByUsername("authuser").
		Return(&database.PublicUser{
			ID:           database.UserID(7),
			Username:     "authuser",
			CreatedAt:    time.Now().Unix(),
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()

	authService := InitDefaultAuthService(mockRepo)
	_, err = authService.AuthenticateUser("authuser", wrongPassword)
	require.Error(t, err)
}

func TestAuthenticateNonexistentUser(t *testing.T) {
	mockRepo := mockUserRepository.NewMockUserRepository(t)
	mockRepo.EXPECT().
		GetUserByUsername("nonexistent").
		Return(nil, assert.AnError).
		Once()

	authService := InitDefaultAuthService(mockRepo)
	_, err := authService.AuthenticateUser("nonexistent", "anyPassword")
	require.Error(t, err)
}

func TestCreateToken(t *testing.T) {
	userID := database.UserID(123)
	tokenString, err := createToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	// Parse the token to verify its contents
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	})
	require.NoError(t, err)
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", userID), claims["sub"])
}

func TestLoginCreatesValidJWT(t *testing.T) {
	mockRepo := mockUserRepository.NewMockUserRepository(t)

	uid := database.UserID(123)

	password := "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 1)
	mockRepo.EXPECT().
		GetUserByUsername("testuser").
		Return(&database.PublicUser{
			ID:           uid,
			Username:     "testuser",
			PasswordHash: string(hashedPassword),
		}, nil).
		Once()

	service := InitDefaultAuthService(mockRepo)
	tokenString, err := service.AuthenticateUser("testuser", password)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		t.Errorf("Token is invalid: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Could not parse claims")
	}
	sub := claims["sub"]
	if fmt.Sprintf("%v", sub) != fmt.Sprintf("%d", uid) {
		t.Errorf("Expected sub claim %d, got %v", uid, sub)
	}
}
