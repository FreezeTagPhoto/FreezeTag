package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/data"
	"freezetag/backend/pkg/images"
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

type JWTClaims struct {
	Permissions data.Permissions `json:"permissions"`
	jwt.RegisteredClaims
}

type ApiClaims struct {
	UserID      database.UserID  `json:"userId"`
	Permissions data.Permissions `json:"permissions"`
}

type ApiCreateToken struct {
	TokenId     database.TokenID `json:"tokenId"`
	TokenString string           `json:"tokenString,omitempty"`
}

type AuthService interface {
	// add a user with the given username and password, returning the created user. The password will be hashed before being stored.
	AddUser(username string, password string) (*database.PublicUser, error)
	// Ensures that there is a valid login by creating a default admin user if no users exist. Should be called at server startup.
	EnsureLogin() error
	// ChangePassword changes the user's password after verifying the current password. Returns an error if the current password is incorrect.
	ChangePassword(userID database.UserID, currentPassword string, newPassword string) error
	// ForceChangePassword changes the user's password without requiring the current password. Use with caution.
	ForceChangePassword(userID database.UserID, newPassword string) error
	// returns a JWT token if the username and password are correct
	AuthenticateUser(username string, password string) (string, error)
	// validates the userID, permissions, and default JWT claims associated with the provided JWT token
	ValidateJWT(tokenString string) (JWTClaims, error)
	// validates the userID and permissions associated with the provided API token
	ValidateAPIToken(token string) (ApiClaims, error)
	// creates an API token. Returns the Plaintext token. Plaintext token is not stored. A token can only have as many or fewer permissions as the user has.
	// Returns an error if the user does not have the requested permissions.
	CreateAPIToken(userID database.UserID, permissions data.Permissions, expiresAt *time.Time, label string) (ApiCreateToken, error)
	// soft delete an API token, returning an error if the token does not exist or could not be revoked
	RevokeAPIToken(userId database.UserID, tokenID database.TokenID) error
	// permanently delete an API token, returning an error if the token does not exist or could not be deleted. Admin only operation
	AdminRevokeAPIToken(tokenID database.TokenID) error
	// delete an API token, returning an error if the token does not exist or could not be deleted
	DeleteAPIToken(tokenID database.TokenID) error
	// Get the IDs of the tokens associated with a user
	GetUserApiTokenInfo(userID database.UserID) ([]database.ApiTokenInfo, error)
	// GetUserById returns the public user information for a given user ID
	GetUserById(userID database.UserID) (*database.PublicUser, error)
	// GetPublicUser returns the public user information for a given user ID, or an error if the user does not exist
	GetPublicUser(userID database.UserID) (*database.PublicUser, error)
	// List all users in the system
	AllUsers() ([]database.PublicUser, error)
	// Delete a user by ID
	DeleteUser(userID database.UserID) error
	// GrantPermissions grants the specified permissions to the user with the given ID
	GrantPermissions(userID database.UserID, permissions data.Permissions) error
	// RevokePermissions revokes the specified permissions from the user with the given ID
	RevokePermissions(userID database.UserID, permissions data.Permissions) error
	// GetUserPermissions returns the permissions associated with the given user ID
	GetUserPermissions(userID database.UserID) (data.Permissions, error)
	// SetUserProfilePicture sets the profile picture for a user, given the user ID and picture data. Returns an error if the user does not exist or the picture data is invalid.
	SetUserProfilePicture(userID database.UserID, requestID database.UserID, pictureData []byte, filename string) error
	// AdminSetUserProfilePicture sets the profile picture for a user, given the user ID and picture data. Returns an error if the user does not exist or the picture data is invalid.
	AdminSetUserProfilePicture(userID database.UserID, pictureData []byte, filename string) error
	// GetUserProfilePicture returns the profile picture for a user, given the user ID. Returns an error if the user does not exist or does not have a profile picture.
	GetUserProfilePicture(userID database.UserID) (database.ProfilePicture, error)
}

type DefaultAuthService struct {
	userDatabase database.UserDatabase
	imageParser  images.Parser
}

func InitDefaultAuthService(userDb database.UserDatabase, imageParser images.Parser) *DefaultAuthService {
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
		userDatabase: userDb,
		imageParser:  imageParser,
	}
}

func (s *DefaultAuthService) EnsureLogin() error {
	users, err := s.userDatabase.AllUsers()
	if err != nil {
		return err
	}
	if len(users) == 0 {
		log.Printf("[WARN] since there are no users, a user with username 'admin' and password 'admin' is being created. Change this ASAP.")
		user, err := s.AddUser("admin", "admin")
		if err != nil {
			return err
		}
		err = s.userDatabase.EnsureAdmin(user.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DefaultAuthService) AuthenticateUser(username string, password string) (string, error) {

	user, err := s.userDatabase.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", err
	}
	permissions, err := s.userDatabase.GetUserPermissions(user.ID)
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
	return s.userDatabase.AddUser(username, string(hash))
}

func (s *DefaultAuthService) ValidateJWT(tokenString string) (JWTClaims, error) {
	claims := JWTClaims{}
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(JwtSecretKey), nil
	}, jwt.WithValidMethods([]string{JwtSigningMethod.Alg()}))

	if err != nil {
		return JWTClaims{}, err
	}
	return claims, nil
}

func (s *DefaultAuthService) ValidateAPIToken(token string) (ApiClaims, error) {
	tokenHash := hashToken(token)
	userID, err := s.userDatabase.GetApiUserID(tokenHash)
	if err != nil {
		return ApiClaims{}, err
	}
	permissions, err := s.userDatabase.GetApiPermissions(tokenHash)
	if err != nil {
		return ApiClaims{}, err
	}
	return ApiClaims{
		UserID:      userID,
		Permissions: permissions,
	}, nil
}

func (s *DefaultAuthService) ChangePassword(userID database.UserID, currentPassword string, newPassword string) error {
	hash, err := s.userDatabase.GetPasswordHash(userID)
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
	return s.userDatabase.SetUserPassword(userID, string(newHash))
}

func (s *DefaultAuthService) ForceChangePassword(userID database.UserID, newPassword string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}
	return s.userDatabase.SetUserPassword(userID, string(newHash))
}

func createTokenWithPermissions(userID database.UserID, permissions data.Permissions) (string, error) {
	JWTClaims := JWTClaims{
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

func (s *DefaultAuthService) CreateAPIToken(userID database.UserID, permissions data.Permissions, expiresAt *time.Time, label string) (ApiCreateToken, error) {
	plaintextTokenBytes := make([]byte, 32)
	_, err := rand.Read(plaintextTokenBytes)
	if err != nil {
		return ApiCreateToken{}, err
	}
	plaintextToken := base64.StdEncoding.EncodeToString(plaintextTokenBytes)
	tokenHash := hashToken(plaintextToken)

	user_permissions, err := s.userDatabase.GetUserPermissions(userID)
	if err != nil {
		return ApiCreateToken{}, err
	}
	if !user_permissions.Contains(permissions) {
		log.Printf("[WARN] User with ID %d attempted to create an API token with permissions that exceed their own", userID)
		return ApiCreateToken{}, fmt.Errorf("invalid permissions requested")
	}

	tokenID, err := s.userDatabase.SaveApiToken(userID, expiresAt, tokenHash, label, permissions)
	if err != nil {
		return ApiCreateToken{}, err
	}
	return ApiCreateToken{
		TokenId:     tokenID,
		TokenString: plaintextToken,
	}, nil
}

// ** these dont need much protection, but the auth service later can add additional checks/sorting/buiness/whatever logic here if needed, and the database layer can focus on just data access**

func (s *DefaultAuthService) RevokeAPIToken(userID database.UserID, tokenID database.TokenID) error {
	return s.userDatabase.RevokeApiToken(userID, tokenID)
}

func (s *DefaultAuthService) AdminRevokeAPIToken(tokenID database.TokenID) error {
	log.Printf("[INFO] revoking API token with ID %d", tokenID)
	return s.userDatabase.AdminRevokeApiToken(tokenID)
}

func (s *DefaultAuthService) DeleteAPIToken(tokenID database.TokenID) error {
	log.Printf("[INFO] Deleting an API token with ID %d", tokenID)
	return s.userDatabase.DeleteApiToken(tokenID)
}

func (s *DefaultAuthService) GetUserApiTokenInfo(userID database.UserID) ([]database.ApiTokenInfo, error) {
	return s.userDatabase.GetUserApiTokenInfo(userID)
}

func (s *DefaultAuthService) GetUserById(userID database.UserID) (*database.PublicUser, error) {
	return s.userDatabase.GetUserById(userID)
}

func (s *DefaultAuthService) GetPublicUser(userID database.UserID) (*database.PublicUser, error) {
	return s.userDatabase.GetUserById(userID)
}

func (s *DefaultAuthService) AllUsers() ([]database.PublicUser, error) {
	return s.userDatabase.AllUsers()
}

func (s *DefaultAuthService) DeleteUser(userID database.UserID) error {
	return s.userDatabase.DeleteUser(userID)
}

func (s *DefaultAuthService) GrantPermissions(userID database.UserID, permissions data.Permissions) error {
	log.Printf("[INFO] Granting permissions %v to user with ID %d", permissions, userID)
	return s.userDatabase.GrantUserPermissions(userID, permissions)
}

func (s *DefaultAuthService) RevokePermissions(userID database.UserID, permissions data.Permissions) error {
	log.Printf("[INFO] Revoking permissions %v from user with ID %d", permissions, userID)
	return s.userDatabase.RevokeUserPermissions(userID, permissions)
}

func (s *DefaultAuthService) GetUserPermissions(userID database.UserID) (data.Permissions, error) {
	return s.userDatabase.GetUserPermissions(userID)
}

func (s *DefaultAuthService) AdminSetUserProfilePicture(userID database.UserID, pictureData []byte, filename string) error {
	log.Printf("[INFO] Admin setting profile picture for user with ID %d", userID)
	return s.setProfilePicture(userID, pictureData, filename)
}

func (s *DefaultAuthService) SetUserProfilePicture(userID database.UserID, requestID database.UserID, pictureData []byte, filename string) error {
	if userID != requestID {
		log.Printf("[WARN] User with ID %d attempted to set profile picture for user with ID %d", requestID, userID)
		return fmt.Errorf("users can only set their own profile picture")
	}
	return s.setProfilePicture(userID, pictureData, filename)
}

func (s *DefaultAuthService) setProfilePicture(userID database.UserID, pictureData []byte, filename string) error {
	// validate picture data by attempting to parse it
	data, err := s.imageParser.ParseImage(filename, pictureData)

	if err != nil {
		return fmt.Errorf("invalid picture data: %w", err)
	}
	profilePicture, err := images.CreateProfilePicture(data)
	if err != nil {
		return fmt.Errorf("could not create profile picture: %w", err)
	}
	return s.userDatabase.SetUserProfilePicture(userID, profilePicture)
}

func (s *DefaultAuthService) GetUserProfilePicture(userID database.UserID) (database.ProfilePicture, error) {
	return s.userDatabase.GetUserProfilePicture(userID)
}

func hashToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}
