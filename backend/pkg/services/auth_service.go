package services

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TODO : Move JWT config to .prop file
var ( 
	JwtSigningMethod = jwt.SigningMethodHS256
	JwtSecretKey = "CHANGEME"
	JwtExpirationHours = time.Duration(24) * time.Hour
)

type AuthService interface {
	AddUser(username string, password string) (*database.PublicUser, error)
	AuthenticateUser(username string, password string) (string, error)
	ValidateJWT(tokenString string) (bool, error)
}

type DefaultAuthService struct {
	userRepo repositories.UserRepository
}

func InitAuthService(userRepo repositories.UserRepository) *DefaultAuthService {
	return &DefaultAuthService{
		userRepo: userRepo,
	}
}

func (s *DefaultAuthService) AuthenticateUser(username string, password string) (string, error) {

	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", err
	}
	return createToken(user.ID)
}

func (s *DefaultAuthService) AddUser(username string, password string) (*database.PublicUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.userRepo.AddUser(username, string(hash))
}

func (s *DefaultAuthService) ValidateJWT(tokenString string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func createToken(userID database.UserID) (string, error) {
	claims := jwt.NewWithClaims(JwtSigningMethod, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(JwtExpirationHours).Unix(),
	})
	tokenString, err := claims.SignedString([]byte(JwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

