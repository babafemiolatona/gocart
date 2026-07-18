package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order *models.Order) error
	CreateOrderTx(tx *gorm.DB, order *models.Order) error
	GetOrderByID(id uint) (*models.Order, error)
	GetOrdersByUserID(userID uint) ([]models.Order, error)
	UpdateOrderStatus(orderID uint, status models.OrderStatus) error
	WithTransaction(fn func(tx *gorm.DB) error) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) CreateOrderTx(tx *gorm.DB, order *models.Order) error {
	return tx.Create(order).Error
}

func (r *orderRepository) GetOrderByID(id uint) (*models.Order, error) {
	var order models.Order

	err := r.db.
		Preload("Items").
		First(&order, id).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) GetOrdersByUserID(userID uint) ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) UpdateOrderStatus(orderID uint, status models.OrderStatus) error {
	result := r.db.
		Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *orderRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
