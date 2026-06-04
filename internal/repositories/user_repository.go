package repositories

import (
	"errors"
	"gocart/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	ExistsByEmail(email string) (bool, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	if result := r.db.Create(user); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}

	if result := r.db.Where("email = ?", email).First(user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return user, nil
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64

	err := r.db.Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	user := &models.User{}

	if result := r.db.First(user, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return user, nil
}

func (r *userRepository) Update(user *models.User) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("User not found")
	}

	return nil
}

func (r *userRepository) Delete(id uint) error {
	result := r.db.Delete(&models.User{}, id)

	if result.RowsAffected == 0 {
		return errors.New("User not found")
	}
	return result.Error
}
