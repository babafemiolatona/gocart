package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetByID(id uint) (*models.User, error)
	CountAll() (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	user := &models.User{}

	if err := r.db.
		Preload("Merchant").
		First(user, id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) CountAll() (int64, error) {
	var count int64

	if err := r.db.
		Model(&models.User{}).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
