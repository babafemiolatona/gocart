package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *models.Category) error
	GetByID(id uint) (*models.Category, error)
	GetAll() ([]models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id uint) (*models.Category, error) {
	category := &models.Category{}

	if result := r.db.First(category, id); result.Error != nil {
		return nil, result.Error
	}
	return category, nil
}

func (r *categoryRepository) GetAll() ([]models.Category, error) {
	var categories []models.Category

	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) Update(category *models.Category) error {
	result := r.db.Model(&models.Category{}).
		Where("id = ?", category.ID).
		Updates(category)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *categoryRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Category{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
