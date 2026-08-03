package repositories

import (
	"errors"

	"gocart/internal/dto"
	"gocart/internal/models"
	"gocart/internal/query"

	"gorm.io/gorm"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uint) (*models.Product, error)
	GetAll(
		query *dto.PaginationQuery,
		filters *query.ProductFilters,
	) ([]models.Product, int64, error)
	Update(id uint, values map[string]interface{}) error
	Delete(id uint) error
	IncrementStockTx(tx *gorm.DB, id uint, qty int) error
	DecrementStockTx(tx *gorm.DB, id uint, qty int) error
	GetBySku(sku string) (*models.Product, error)
	CountByMerchant(merchantID uint) (int64, error)
	CountLowStockByMerchant(merchantID uint, threshold int) (int64, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetByID(id uint) (*models.Product, error) {
	product := &models.Product{}

	if err := r.db.
		Preload("Category").
		Preload("Images").
		First(product, id).Error; err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) GetAll(
	query *dto.PaginationQuery,
	filters *query.ProductFilters,
) ([]models.Product, int64, error) {

	var (
		products []models.Product
		total    int64
	)

	db := r.db.Model(&models.Product{})

	if filters != nil {

		if filters.MerchantID > 0 {
			db = db.Where("merchant_id = ?", filters.MerchantID)
		}

		if filters.CategoryID > 0 {
			db = db.Where("category_id = ?", filters.CategoryID)
		}

		if filters.MinPrice > 0 {
			db = db.Where("price >= ?", filters.MinPrice)
		}

		if filters.MaxPrice > 0 {
			db = db.Where("price <= ?", filters.MaxPrice)
		}

		if filters.InStock != nil && *filters.InStock {
			db = db.Where("stock > 0")
		}

		if filters.SearchQuery != "" {
			db = db.Where(
				"name ILIKE ? OR description ILIKE ?",
				"%"+filters.SearchQuery+"%",
				"%"+filters.SearchQuery+"%",
			)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize

	allowedSorts := map[string]bool{
		"id":         true,
		"name":       true,
		"price":      true,
		"stock":      true,
		"created_at": true,
	}

	sortField := "created_at"
	if allowedSorts[query.Sort] {
		sortField = query.Sort
	}

	db = db.
		Preload("Category").
		Preload("Images").
		Offset(offset).
		Limit(query.PageSize).
		Order(sortField + " " + query.Order)

	if err := db.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(id uint, values map[string]interface{}) error {
	return r.db.
		Model(&models.Product{}).
		Where("id = ?", id).
		Updates(values).Error
}

func (r *productRepository) IncrementStockTx(
	tx *gorm.DB,
	id uint,
	qty int,
) error {
	result := tx.
		Model(&models.Product{}).
		Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", qty))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *productRepository) DecrementStockTx(
	tx *gorm.DB,
	id uint,
	qty int,
) error {
	result := tx.
		Model(&models.Product{}).
		Where("id = ? AND stock >= ?", id, qty).
		Update("stock", gorm.Expr("stock - ?", qty))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		var exists int64

		if err := tx.
			Model(&models.Product{}).
			Where("id = ?", id).
			Count(&exists).Error; err != nil {
			return err
		}

		if exists == 0 {
			return gorm.ErrRecordNotFound
		}

		return ErrInsufficientStock
	}

	return nil
}

func (r *productRepository) Delete(id uint) error {

	result := r.db.Delete(&models.Product{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *productRepository) GetBySku(sku string) (*models.Product, error) {

	product := &models.Product{}

	if err := r.db.
		Where("sku = ?", sku).
		First(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) CountByMerchant(merchantID uint) (int64, error) {
	var count int64

	if err := r.db.
		Model(&models.Product{}).
		Where("merchant_id = ?", merchantID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *productRepository) CountLowStockByMerchant(merchantID uint, threshold int) (int64, error) {

	var count int64

	if err := r.db.
		Model(&models.Product{}).
		Where("merchant_id = ? AND stock <= ?", merchantID, threshold).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
