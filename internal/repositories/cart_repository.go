package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart *models.Cart) error
	GetByUserID(userID uint) (*models.Cart, error)
	GetWithItems(userID uint) (*models.Cart, error)
	AddItem(item *models.CartItem) error
	UpdateItem(item *models.CartItem) error
	RemoveItem(cartItemID uint) error
	UpdateCartTotal(cartID uint, total float64, itemCount int) error
	ClearCart(cartID uint) error
	ClearCartTx(tx *gorm.DB, cartID uint) error
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) Create(cart *models.Cart) error {
	return r.db.Create(cart).Error
}

func (r *cartRepository) GetByUserID(userID uint) (*models.Cart, error) {
	var cart models.Cart
	if err := r.db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetWithItems(userID uint) (*models.Cart, error) {
	var cart models.Cart

	if err := r.db.
		Preload("Items").
		Preload("Items.Product").
		Where("user_id = ?", userID).
		First(&cart).Error; err != nil {
		return nil, err
	}

	return &cart, nil
}

func (r *cartRepository) AddItem(item *models.CartItem) error {
	return r.db.Create(item).Error
}

func (r *cartRepository) UpdateItem(item *models.CartItem) error {
	result := r.db.
		Model(&models.CartItem{}).
		Where("id = ?", item.ID).
		Updates(map[string]interface{}{
			"quantity": item.Quantity,
			"price":    item.Price,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *cartRepository) RemoveItem(cartItemID uint) error {
	result := r.db.Delete(&models.CartItem{}, cartItemID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *cartRepository) UpdateCartTotal(cartID uint, total float64, itemCount int) error {
	return r.db.Model(&models.Cart{}).
		Where("id = ?", cartID).
		Updates(map[string]interface{}{
			"total":      total,
			"item_count": itemCount,
		}).Error
}

func (r *cartRepository) ClearCart(cartID uint) error {
	return r.db.Where("cart_id = ?", cartID).
		Delete(&models.CartItem{}).Error
}

func (r *cartRepository) ClearCartTx(tx *gorm.DB, cartID uint) error {
	return tx.
		Where("cart_id = ?", cartID).
		Delete(&models.CartItem{}).
		Error
}
