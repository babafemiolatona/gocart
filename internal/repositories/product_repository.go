package repositories

import (
	"gocart/internal/models"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uint) (*models.Product, error)
	GetAll(query *models.PaginationQuery, filters *models.ProductFilters) ([]models.Product, int64, error)
	Update(product *models.Product) error
	Delete(id uint) error
	GetBySku(sku string) (*models.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	if result := r.db.Create(product); result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *productRepository) GetByID(id uint) (*models.Product, error) {
	product := &models.Product{}

	if err := r.db.Preload("Category").First(product, id).Error; err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) GetAll(query *models.PaginationQuery, filters *models.ProductFilters) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	db := r.db.Model(&models.Product{})

	if filters != nil {
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

	db = db.Preload("Category")

	offset := (query.Page - 1) * query.PageSize
	db = db.Offset(offset).Limit(query.PageSize)

	allowedSorts := map[string]bool{
		"id":         true,
		"name":       true,
		"price":      true,
		"created_at": true,
		"stock":      true,
	}

	sortField := "created_at"
	if allowedSorts[query.Sort] {
		sortField = query.Sort
	}

	db = db.Order(sortField + " " + query.Order)

	if err := db.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(product *models.Product) error {
	result := r.db.Model(&models.Product{}).
		Where("id = ?", product.ID).
		Updates(product)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *productRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Product{}, id)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *productRepository) GetBySku(sku string) (*models.Product, error) {
	product := &models.Product{}

	if err := r.db.Where("sku = ?", sku).First(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}
