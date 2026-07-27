package models

import (
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
	Merchant  *Merchant `gorm:"foreignKey:UserID" json:"merchant,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func IsValidRole(role Role) bool {
	return role == RoleCustomer || role == RoleAdmin
}
