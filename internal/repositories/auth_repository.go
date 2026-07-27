package repositories

import (
	"errors"
	"gocart/internal/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByEmailOrUsername(identifier string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	GetByIDTx(tx *gorm.DB, id uint) (*models.User, error)
	ExistsByEmail(email string) (bool, error)
	Update(user *models.User) error
	UpdateTx(tx *gorm.DB, user *models.User) error
	Delete(id uint) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}

	if err := r.db.
		Where("email = ?", email).
		First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *authRepository) GetByEmailOrUsername(identifier string) (*models.User, error) {
	user := &models.User{}

	if err := r.db.
		Where(
			"LOWER(email) = LOWER(?) OR LOWER(username) = LOWER(?)",
			identifier,
			identifier,
		).
		First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *authRepository) ExistsByEmail(email string) (bool, error) {
	var count int64

	if err := r.db.
		Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *authRepository) GetByID(id uint) (*models.User, error) {
	user := &models.User{}

	if err := r.db.
		First(user, id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *authRepository) GetByIDTx(tx *gorm.DB, id uint) (*models.User, error) {
	var user models.User

	if err := tx.
		First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *authRepository) Update(user *models.User) error {
	result := r.db.
		Model(&models.User{}).
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

func (r *authRepository) UpdateTx(tx *gorm.DB, user *models.User) error {
	result := tx.
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) Delete(id uint) error {
	result := r.db.Delete(&models.User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("User not found")
	}
	return result.Error
}
