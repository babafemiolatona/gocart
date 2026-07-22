package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *models.Payment) error
	CreateTx(tx *gorm.DB, payment *models.Payment) error
	GetByReference(reference string) (*models.Payment, error)
	GetByOrderID(orderID uint) (*models.Payment, error)
	UpdateStatus(reference string, status models.PaymentStatus) error
	UpdateStatusTx(tx *gorm.DB, reference string, status models.PaymentStatus) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) CreateTx(tx *gorm.DB, payment *models.Payment) error {
	return tx.Create(payment).Error
}

func (r *paymentRepository) GetByReference(reference string) (*models.Payment, error) {
	var payment models.Payment

	if err := r.db.
		Where("reference = ?", reference).
		First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByOrderID(orderID uint) (*models.Payment, error) {
	var payment models.Payment

	if err := r.db.
		Where("order_id = ?", orderID).
		First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(reference string, status models.PaymentStatus) error {
	return r.db.Model(&models.Payment{}).
		Where("reference = ?", reference).
		Update("status", status).Error
}

func (r *paymentRepository) UpdateStatusTx(tx *gorm.DB, reference string, status models.PaymentStatus) error {
	return tx.Model(&models.Payment{}).
		Where("reference = ?", reference).
		Update("status", status).Error
}
