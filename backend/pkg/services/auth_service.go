package services

import (
	"freezetag/backend/pkg/database"
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
	AddUser(username string, passwordHash string) error
	AuthenticateUser(username string, password string) (string, error)
}

type DefaultAuthService struct {
	userDB database.UserDatabase
}

func InitAuthService(userDB database.UserDatabase) *DefaultAuthService {
	return &DefaultAuthService{
		userDB: userDB,
	}
}

func (s *DefaultAuthService) AuthenticateUser(username string, password string) (string, error) {

	userID, err := s.userDB.GetUserIDByUsername(username)
	if err != nil {
		return "", err
	}
	storedHash, err := s.userDB.GetUserPasswordHashByID(userID)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		return "", err
	}
	return createToken(userID)
}

func (s *DefaultAuthService) AddUser(username string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userDB.AddUser(username, string(hash))
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