package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type productImageRepository struct {
	db *gorm.DB
}

type ProductImageRepository interface {
	Create(image *models.ProductImage) error
	CreateMany(images []models.ProductImage) error
	GetByProductID(productID uint) ([]models.ProductImage, error)
	Delete(id uint) error
}

func NewProductImageRepository(db *gorm.DB) ProductImageRepository {
	return &productImageRepository{
		db: db,
	}
}

func (r *productImageRepository) Create(image *models.ProductImage) error {
	return r.db.Create(image).Error
}

func (r *productImageRepository) CreateMany(images []models.ProductImage) error {
	return r.db.Create(&images).Error
}

func (r *productImageRepository) GetByProductID(productID uint) ([]models.ProductImage, error) {
	var images []models.ProductImage

	err := r.db.
		Where("product_id = ?", productID).
		Find(&images).Error

	return images, err
}

func (r *productImageRepository) Delete(id uint) error {
	return r.db.Delete(&models.ProductImage{}, id).Error
}
