package services

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gocart/internal/config"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	authRepo repositories.AuthRepository
	config   *config.Config
}

func NewAuthService(authRepo repositories.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		config:   cfg,
	}
}

func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.UserResponse, error) {

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := req.Validate(); err != nil {
		return nil, apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		)
	}

	exists, err := s.authRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"check_user_failed",
			"failed to check if user already exists",
			err,
		)
	}

	if exists {
		return nil, apperrors.New(
			http.StatusConflict,
			"user_exists",
			"user already exists",
			nil,
		)
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Role:      models.RoleCustomer,
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"hash_password_failed",
			"failed to hash password",
			err,
		)
	}

	if err := s.authRepo.Create(user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				"user_exists",
				"user already exists",
				nil,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"create_user_failed",
			"failed to create user",
			err,
		)
	}

	return mapper.ToUserResponse(user), nil
}

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	identifier := strings.ToLower(strings.TrimSpace(req.UsernameOrEmail))

	user, err := s.authRepo.GetByEmailOrUsername(identifier)
	if err != nil {
		return nil, apperrors.New(
			http.StatusUnauthorized,
			"invalid_credentials",
			"invalid username/email or password",
			nil,
		)
	}

	if !user.VerifyPassword(req.Password) {
		return nil, apperrors.New(
			http.StatusUnauthorized,
			"invalid_credentials",
			"invalid username/email or password",
			nil,
		)
	}

	token, expiresAt, err := s.GenerateToken(user)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"generate_token_failed",
			"failed to generate token",
			err,
		)
	}

	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

type CustomClaims struct {
	// ID    uint   `json:"id"`
	// Email string `json:"email"`
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateToken(user *models.User) (string, int64, error) {
	expiresAt := time.Now().Add(s.config.JWTExpiry)

	claims := CustomClaims{
		// ID:    user.ID,
		// Email: user.Email,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.JWTSecret))

	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresAt.Unix(), nil
}

func (s *AuthService) VerifyToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("invalid signing method")
			}
			return []byte(s.config.JWTSecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) GetJWTSecret() string {
	return s.config.JWTSecret
}
