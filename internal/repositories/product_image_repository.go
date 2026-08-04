package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type productImageRepository struct {
	db *gorm.DB
}

type ProductImageRepository interface {
	CreateMany(images []models.ProductImage) error
}

func NewProductImageRepository(db *gorm.DB) ProductImageRepository {
	return &productImageRepository{
		db: db,
	}
}

func (r *productImageRepository) CreateMany(images []models.ProductImage) error {
	return r.db.Create(&images).Error
}
