package repositories

import (
	"gocart/internal/dto"
	"gocart/internal/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrderTx(tx *gorm.DB, order *models.Order) error
	GetOrderByID(id uint) (*models.Order, error)
	GetByUserIDAndIdempotencyKey(userID uint, key string) (*models.Order, error)
	GetOrdersByUserID(userID uint, p *dto.PaginationQuery) ([]models.Order, int64, error)
	UpdateOrderStatus(orderID uint, status models.OrderStatus) error
	UpdateOrderStatusTx(tx *gorm.DB, orderID uint, status models.OrderStatus) error
	WithTransaction(fn func(tx *gorm.DB) error) error
	GetOrdersByMerchantID(merchantID uint, p *dto.PaginationQuery) ([]models.Order, int64, error)
	GetMerchantOrderByID(merchantID uint, orderID uint) (*models.Order, error)

	CountByMerchant(merchantID uint) (int64, error)
	CountByMerchantAndStatus(merchantID uint, status models.OrderStatus) (int64, error)
	SumRevenueByMerchant(merchantID uint) (int64, error)
	GetRecentOrdersByMerchant(merchantID uint, limit int) ([]models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
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

func (r *orderRepository) GetByUserIDAndIdempotencyKey(userID uint, key string) (*models.Order, error) {
	var order models.Order

	err := r.db.
		Preload("Items").
		Where("user_id = ?", userID).
		Where("idempotency_key = ?", key).
		First(&order).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) GetOrdersByUserID(userID uint, p *dto.PaginationQuery) ([]models.Order, int64, error) {

	var orders []models.Order
	var total int64

	if err := r.db.
		Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "DESC"
	if p.Order == "asc" {
		order = "ASC"
	}

	offset := (p.Page - 1) * p.PageSize

	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at " + order).
		Offset(offset).
		Limit(p.PageSize).
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
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

func (r *orderRepository) UpdateOrderStatusTx(tx *gorm.DB, orderID uint, status models.OrderStatus) error {
	result := tx.
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

func (r *orderRepository) GetOrdersByMerchantID(
	merchantID uint,
	p *dto.PaginationQuery,
) ([]models.Order, int64, error) {

	var orders []models.Order
	var total int64

	baseQuery := r.db.
		Model(&models.Order{}).
		Distinct("orders.*").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.merchant_id = ?", merchantID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "DESC"
	if p.Order == "asc" {
		order = "ASC"
	}

	offset := (p.Page - 1) * p.PageSize

	err := baseQuery.
		Order("orders.created_at " + order).
		Offset(offset).
		Limit(p.PageSize).
		Preload("User").
		Preload("Items").
		Find(&orders).Error

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) GetMerchantOrderByID(
	merchantID uint,
	orderID uint,
) (*models.Order, error) {

	var order models.Order

	err := r.db.
		Distinct("orders.*").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.id = ?", orderID).
		Where("products.merchant_id = ?", merchantID).
		Preload("User").
		Preload("Items").
		First(&order).Error

	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) CountByMerchant(merchantID uint) (int64, error) {
	var count int64

	err := r.db.
		Model(&models.Order{}).
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.merchant_id = ?", merchantID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *orderRepository) CountByMerchantAndStatus(merchantID uint, status models.OrderStatus) (int64, error) {

	var count int64

	err := r.db.
		Model(&models.Order{}).
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.merchant_id = ?", merchantID).
		Where("orders.status = ?", status).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *orderRepository) SumRevenueByMerchant(merchantID uint) (int64, error) {

	var total int64

	err := r.db.
		Model(&models.Order{}).
		Select("COALESCE(SUM(orders.total), 0)").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.merchant_id = ?", merchantID).
		Where("orders.status = ?", models.OrderStatusDelivered).
		Scan(&total).Error

	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *orderRepository) GetRecentOrdersByMerchant(merchantID uint, limit int) ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		Distinct("orders.*").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.merchant_id = ?", merchantID).
		Order("orders.created_at DESC").
		Limit(limit).
		Preload("User").
		Preload("Items").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}
