package models

import (
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Role      Role      `gorm:"not null" json:"role"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  string    `gorm:"not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(password),
	)
	return err == nil
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

func IsValidRole(role Role) bool {
	return role == RoleCustomer || role == RoleAdmin
}
