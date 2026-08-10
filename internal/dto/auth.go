package dto

import (
	"errors"
	"gocart/internal/models"
	"regexp"
	"time"
)

type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
}

type UserResponse struct {
	ID        uint        `json:"id"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Role      models.Role `json:"role"`
	CreatedAt time.Time   `json:"created_at"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func (r *ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == "" {
		return errors.New("current password is required")
	}
	if r.NewPassword == "" || len(r.NewPassword) < 6 {
		return errors.New("new password must be at least 6 characters")
	}
	if r.NewPassword != r.ConfirmPassword {
		return errors.New("passwords do not match")
	}
	if r.CurrentPassword == r.NewPassword {
		return errors.New("new password must be different from the current password")
	}
	return nil
}

func ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	if !matched {
		return errors.New("invalid email format")
	}
	return nil
}

func (r *RegisterRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}
	if r.Password == "" || len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if r.Password != r.ConfirmPassword {
		return errors.New("passwords do not match")
	}
	if r.FirstName == "" || r.LastName == "" {
		return errors.New("first name and last name are required")
	}
	return nil
}
