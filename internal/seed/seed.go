package seed

import (
	"errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(userRepo repositories.UserRepository) error {
	_, err := userRepo.GetByEmail("admin@gocart.com")
	if err == nil {
		log.Println("Admin user already exists, skipping seeding")
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("admin123"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	adminUser := &models.User{
		Username:  "admin",
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@gocart.com",
		Password:  string(hashedPassword),
		Role:      models.RoleAdmin,
	}

	if err := userRepo.Create(adminUser); err != nil {
		return err
	}

	log.Println("Admin user seeded successfully")
	return nil
}
