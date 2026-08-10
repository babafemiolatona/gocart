package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type MerchantRepository interface {
	Create(merchant *models.Merchant) error
	GetByID(id uint) (*models.Merchant, error)
	GetByUserID(userID uint) (*models.Merchant, error)
	Update(merchant *models.Merchant) error
	CountAll() (int64, error)
}

type merchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) MerchantRepository {
	return &merchantRepository{
		db: db,
	}
}

func (r *merchantRepository) Create(merchant *models.Merchant) error {
	return r.db.Create(merchant).Error
}

func (r *merchantRepository) GetByID(id uint) (*models.Merchant, error) {
	var merchant models.Merchant

	if err := r.db.First(&merchant, id).Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

func (r *merchantRepository) GetByUserID(userID uint) (*models.Merchant, error) {
	var merchant models.Merchant

	if err := r.db.
		Where("user_id = ?", userID).
		First(&merchant).Error; err != nil {
		return nil, err
	}

	return &merchant, nil
}

func (r *merchantRepository) Update(merchant *models.Merchant) error {
	return r.db.Save(merchant).Error
}

func (r *merchantRepository) CountAll() (int64, error) {
	var count int64

	if err := r.db.
		Model(&models.Merchant{}).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
