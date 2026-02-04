package services

import (
	"crypto/rand"
	"encoding/base64"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

var (
	JwtSigningMethod   = jwt.SigningMethodHS256
	JwtSecretKey       = ""
	JwtExpirationHours = time.Duration(24) * time.Hour
)

type AuthService interface {
	AddUser(username string, password string) (*database.PublicUser, error)
	AuthenticateUser(username string, password string) (string, error)
	ValidateJWT(tokenString string) (jwt.MapClaims, error)
}

type DefaultAuthService struct {
	userRepo repositories.UserRepository
}

func InitDefaultAuthService(userRepo repositories.UserRepository) *DefaultAuthService {
	key, exists := os.LookupEnv("JWT_SECRET_KEY")
	if !exists || key == "" {
		log.Printf("[WARN] JWT_SECRET_KEY in .env file was not found or was empty, defaulting to random bytes")
		randomBytes := make([]byte, 32)
		_, err := rand.Read(randomBytes)
		if err != nil {
			panic(err)
		}
		JwtSecretKey = base64.StdEncoding.EncodeToString(randomBytes)
	} else {
		JwtSecretKey = key
	}
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

func (s *DefaultAuthService) ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	}, jwt.WithValidMethods([]string{JwtSigningMethod.Alg()}))

	if err != nil {
		return nil, err
	}
	return claims, nil
}

func createToken(userID database.UserID) (string, error) {
	claims := jwt.NewWithClaims(JwtSigningMethod, jwt.MapClaims{
		"sub": strconv.FormatInt(int64(userID), 10),
		"exp": time.Now().Add(JwtExpirationHours).Unix(),
	})
	tokenString, err := claims.SignedString([]byte(JwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
