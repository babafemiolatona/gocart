package repositories

import (
	"fmt"

	"gocart/internal/config"
	"gocart/internal/logger"
	"gocart/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger:         gormlogger.Default.LogMode(gormlogger.Info),
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Log.Info().Msg("database connection successful")

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.ProductImage{},
		&models.Payment{},
		&models.Merchant{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Log.Info().Msg("database migration successful")
	return db, nil
}
