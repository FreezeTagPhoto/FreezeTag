package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
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
	bcryptCost         = bcrypt.DefaultCost
)

type Claims struct {
	Permissions data.Permissions `json:"permissions"`
	jwt.RegisteredClaims
}

type AuthService interface {
	AddUser(username string, password string) (*database.PublicUser, error)
	EnsureLogin() error
	ChangePassword(userID database.UserID, currentPassword string, newPassword string) error
	ForceChangePassword(userID database.UserID, newPassword string) error
	AuthenticateUser(username string, password string) (string, error)
	ValidateJWT(tokenString string) (Claims, error)
	ValidateAPIToken(token string) (data.Permissions, error)
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

func (s *DefaultAuthService) EnsureLogin() error {
	users, err := s.userRepo.ListAllUsers()
	if err != nil {
		return err
	}
	if len(users) == 0 {
		log.Printf("[WARN] since there are no users, a user with username 'admin' and password 'admin' is being created. Change this ASAP.")
		user, err := s.AddUser("admin", "admin")
		if err != nil {
			return err
		}
		err = s.userRepo.GrantAdminPermissions(user.ID)
		if err != nil {
			return err
		}
	}
	return nil
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
	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		return "", err
	}

	return createTokenWithPermissions(user.ID, permissions)
}

func (s *DefaultAuthService) AddUser(username string, password string) (*database.PublicUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}
	return s.userRepo.AddUser(username, string(hash))
}

func (s *DefaultAuthService) ValidateJWT(tokenString string) (Claims, error) {
	claims := Claims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	}, jwt.WithValidMethods([]string{JwtSigningMethod.Alg()}))

	if err != nil {
		return Claims{}, err
	}
	return claims, nil
}

func (s *DefaultAuthService) ValidateAPIToken(token string) (data.Permissions, error) {
	tokenHash := hashToken(token)
	permissions, err := s.userRepo.GetApiPermissions(tokenHash)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (s *DefaultAuthService) ChangePassword(userID database.UserID, currentPassword string, newPassword string) error {
	hash, err := s.userRepo.GetUserPasswordHash(userID)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword))
	if err != nil {
		return err
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	return s.userRepo.ChangePassword(userID, string(newHash))
}

func (s *DefaultAuthService) ForceChangePassword(userID database.UserID, newPassword string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	return s.userRepo.ChangePassword(userID, string(newHash))
}

func createTokenWithPermissions(userID database.UserID, permissions data.Permissions) (string, error) {
	JWTClaims := Claims{
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(int64(userID), 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(JwtExpirationHours)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(JwtSigningMethod, JWTClaims)
	return token.SignedString([]byte(JwtSecretKey))
}

func hashToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}
