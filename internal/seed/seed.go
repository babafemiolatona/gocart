package seed

import (
	"errors"
	"gocart/internal/logger"
	"gocart/internal/models"
	"gocart/internal/repositories"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(authRepo repositories.AuthRepository) error {
	_, err := authRepo.GetByEmail("admin@gocart.com")
	if err == nil {
		logger.Log.Info().Msg("admin user already exists, skipping seeding")
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

	if err := authRepo.Create(adminUser); err != nil {
		return err
	}

	logger.Log.Info().Msg("admin user seeded successfully")
	return nil
}
