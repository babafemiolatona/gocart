package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gocart/internal/config"
	"gocart/internal/models"
	"gocart/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
)

type UserService struct {
	userRepo repositories.UserRepository
	config   *config.Config
}

func NewUserService(userRepo repositories.UserRepository, cfg *config.Config) *UserService {
	return &UserService{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := req.Validate(); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err == nil && exists {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}

	if exists {
		return nil, errors.New("user already exists")
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Role:      models.RoleCustomer, // Default to customer role
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, errors.New("failed to hash password")
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *UserService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.VerifyPassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	token, expiresAt, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.AuthResponse{
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

func (s *UserService) GenerateToken(user *models.User) (string, int64, error) {
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

func (s *UserService) VerifyToken(tokenStr string) (*CustomClaims, error) {
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

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) GetJWTSecret() string {
	return s.config.JWTSecret
}
