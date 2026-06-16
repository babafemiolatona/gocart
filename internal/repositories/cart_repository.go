package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type CartRepository interface {
	Create(cart *models.Cart) error
	// GetOrCreate(userID uint) (*models.Cart, error)
	GetByUserID(userID uint) (*models.Cart, error)
	GetWithItems(userID uint) (*models.Cart, error)
	AddItem(item *models.CartItem) error
	// GetItem(cartID uint, productID uint) (*models.CartItem, error)
	UpdateItem(item *models.CartItem) error
	RemoveItem(cartItemID uint) error
	UpdateCartTotal(cartID uint, total float64, itemCount int) error
	ClearCart(cartID uint) error
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

// func (r *cartRepository) GetOrCreate(userID uint) (*models.Cart, error) {
// 	var cart models.Cart

// 	if err := r.db.
// 		Preload("Items.Product").
// 		Where("user_id = ?", userID).
// 		FirstOrCreate(&cart, models.Cart{UserID: userID}).Error; err != nil {
// 		return nil, err
// 	}

// 	return &cart, nil
// }

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

// func (r *cartRepository) GetItem(cartID uint, productID uint) (*models.CartItem, error) {
// 	var item models.CartItem

// 	if err := r.db.
// 		Where("cart_id = ? AND product_id = ?", cartID, productID).
// 		First(&item).Error; err != nil {
// 		return nil, err
// 	}
// 	return &item, nil
// }

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

	return nil
}

// func (r *cartRepository) UpdateCartTotal(cartID uint) error {
// 	var total float64
// 	var itemCount int

// 	if err := r.db.Model(&models.CartItem{}).
// 		Select("COALESCE(SUM(quantity * price), 0), COALESCE(SUM(quantity), 0)").
// 		Where("cart_id = ?", cartID).
// 		Row().
// 		Scan(&total, &itemCount); err != nil {
// 		return err
// 	}

// 	result := r.db.Model(&models.Cart{}).
// 		Where("id = ?", cartID).
// 		Updates(map[string]interface{}{
// 			"total":      total,
// 			"item_count": itemCount,
// 		})

// 	if result.Error != nil {
// 		return result.Error
// 	}

// 	if result.RowsAffected == 0 {
// 		return gorm.ErrRecordNotFound
// 	}

// 	return nil
// }

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

// func (r *cartRepository) ClearCart(cartID uint) error {
// 	return r.db.Transaction(func(tx *gorm.DB) error {
// 		if err := tx.Where("cart_id = ?", cartID).Delete(&models.CartItem{}).Error; err != nil {
// 			return err
// 		}

// 		result := tx.Model(&models.Cart{}).
// 			Where("id = ?", cartID).
// 			Updates(map[string]interface{}{
// 				"total":      0,
// 				"item_count": 0,
// 			})

// 		if result.Error != nil {
// 			return result.Error
// 		}

// 		if result.RowsAffected == 0 {
// 			return gorm.ErrRecordNotFound
// 		}

// 		return nil
// 	})
// }
