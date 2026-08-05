package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type MerchantRepository interface {
	Create(merchant *models.Merchant) error
	CreateTx(tx *gorm.DB, merchant *models.Merchant) error
	GetByID(id uint) (*models.Merchant, error)
	GetByUserID(userID uint) (*models.Merchant, error)
	Update(merchant *models.Merchant) error
	WithTransaction(fn func(tx *gorm.DB) error) error
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

func (r *merchantRepository) CreateTx(tx *gorm.DB, merchant *models.Merchant) error {
	return tx.Create(merchant).Error
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

func (r *merchantRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
