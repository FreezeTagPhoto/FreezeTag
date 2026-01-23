package services

import (
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TODO: change this so that its not horrifically insecure
const ( 
	JwtSecretKey = "CHANGEME"
	JwtExpirationHours = time.Duration(24) * time.Hour
)

type AuthService interface {
	AddUser(username string, passwordHash string) (database.UserID, error)
	AuthenticateUser(username string, password string) (string, error)
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
	storedHash, err := s.userRepo.GetUserPasswordHash(user.ID)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		return "", err
	}
	return createToken(user.ID)
}

func (s *DefaultAuthService) AddUser(username string, password string) (database.UserID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	return s.userRepo.AddUser(username, string(hash))
}

func createToken(userID database.UserID) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(JwtExpirationHours).Unix(),
	})
	tokenString, err := claims.SignedString([]byte(JwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}