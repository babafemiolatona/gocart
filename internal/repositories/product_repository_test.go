package repositories

import (
	"errors"
	"testing"

	"gocart/internal/dto"
	"gocart/internal/models"
	"gocart/internal/query"

	"gorm.io/gorm"
)

func TestProductRepository_GetAll_Filters(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)

	otherCategory := &models.Category{Name: "Gadgets", Slug: "gadgets"}
	if err := db.Create(otherCategory).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	third := &models.Product{
		Name:       "Widget",
		Slug:       "widget",
		Sku:        "WDG-001",
		Price:      1500,
		Stock:      3,
		CategoryID: otherCategory.ID,
		MerchantID: fixture.merchant.ID,
	}
	if err := db.Create(third).Error; err != nil {
		t.Fatalf("create third product: %v", err)
	}

	repo := productRepo(db)
	pq := &dto.PaginationQuery{Page: 1, PageSize: 10}

	t.Run("no filters returns all", func(t *testing.T) {
		products, total, err := repo.GetAll(pq, &query.ProductFilters{})
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if total != 3 {
			t.Fatalf("want total 3, got %d", total)
		}
		if len(products) != 3 {
			t.Fatalf("want 3 products, got %d", len(products))
		}
	})

	t.Run("category filter", func(t *testing.T) {
		products, total, err := repo.GetAll(pq, &query.ProductFilters{CategoryID: fixture.category.ID})
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if total != 2 {
			t.Fatalf("want 2 in category, got %d", total)
		}
		for _, p := range products {
			if p.CategoryID != fixture.category.ID {
				t.Fatalf("product %d not in expected category", p.ID)
			}
		}
	})

	t.Run("merchant filter", func(t *testing.T) {
		otherMerchant := &models.Merchant{UserID: 2, BusinessName: "Others"}
		if err := db.Create(otherMerchant).Error; err != nil {
			t.Fatalf("create merchant: %v", err)
		}
		if _, total, err := repo.GetAll(pq, &query.ProductFilters{MerchantID: otherMerchant.ID}); err != nil {
			t.Fatalf("GetAll: %v", err)
		} else if total != 0 {
			t.Fatalf("want 0 for empty merchant, got %d", total)
		}
		if _, total, err := repo.GetAll(pq, &query.ProductFilters{MerchantID: fixture.merchant.ID}); err != nil {
			t.Fatalf("GetAll: %v", err)
		} else if total != 3 {
			t.Fatalf("want 3 for merchant, got %d", total)
		}
	})

	t.Run("price min/max", func(t *testing.T) {
		products, total, err := repo.GetAll(pq, &query.ProductFilters{MinPrice: 2000, MaxPrice: 4000})
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if total != 1 {
			t.Fatalf("want 1 in price range, got %d", total)
		}
		if products[0].Price != 3000 {
			t.Fatalf("want 3000 product, got %d", products[0].Price)
		}
	})

	t.Run("in stock filter", func(t *testing.T) {
		inStock := true
		products, total, err := repo.GetAll(pq, &query.ProductFilters{InStock: &inStock})
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if total != 2 {
			t.Fatalf("want 2 in stock, got %d", total)
		}
		for _, p := range products {
			if p.Stock <= 0 {
				t.Fatalf("product %d should be in stock", p.ID)
			}
		}
	})

	t.Run("pagination and sort by price", func(t *testing.T) {
		products, total, err := repo.GetAll(&dto.PaginationQuery{Page: 1, PageSize: 2, Sort: "price", Order: "desc"}, &query.ProductFilters{})
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if total != 3 {
			t.Fatalf("want total 3, got %d", total)
		}
		if len(products) != 2 {
			t.Fatalf("want 2 products, got %d", len(products))
		}
		if products[0].Price < products[1].Price {
			t.Fatal("want descending price order")
		}
	})
}

func TestProductRepository_StockTransitions(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	repo := productRepo(db)

	product := fixture.productA

	if err := repo.IncrementStock(product.ID, 5); err != nil {
		t.Fatalf("IncrementStock: %v", err)
	}

	if err := repo.DecrementStock(product.ID, 3); err != nil {
		t.Fatalf("DecrementStock: %v", err)
	}

	var reloaded models.Product
	if err := db.First(&reloaded, product.ID).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if reloaded.Stock != 12 {
		t.Fatalf("want stock 12 (10+5-3), got %d", reloaded.Stock)
	}

	t.Run("decrement below stock is insufficient", func(t *testing.T) {
		if err := repo.DecrementStock(product.ID, 999); !errors.Is(err, ErrInsufficientStock) {
			t.Fatalf("want ErrInsufficientStock, got %v", err)
		}
	})

	t.Run("decrement missing product", func(t *testing.T) {
		if err := repo.DecrementStock(99999, 1); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("want ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("increment missing product", func(t *testing.T) {
		if err := repo.IncrementStock(99999, 1); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("want ErrRecordNotFound, got %v", err)
		}
	})
}

func TestProductRepository_Delete(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	repo := productRepo(db)

	if err := db.Create(&models.Product{
		Name:       "Doomed",
		Slug:       "doomed",
		Sku:        "DOOM-001",
		Price:      100,
		Stock:      1,
		CategoryID: fixture.category.ID,
		MerchantID: fixture.merchant.ID,
	}).Error; err != nil {
		t.Fatalf("create doomed product: %v", err)
	}

	doomed := &models.Product{}
	if err := db.Where("sku = ?", "DOOM-001").First(doomed).Error; err != nil {
		t.Fatalf("load doomed product: %v", err)
	}

	if err := repo.Delete(doomed.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := repo.Delete(doomed.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("want ErrRecordNotFound on second delete, got %v", err)
	}
}

func TestProductRepository_Update(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	repo := productRepo(db)

	if err := repo.Update(fixture.productA.ID, map[string]interface{}{"price": 4000}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	var reloaded models.Product
	if err := db.First(&reloaded, fixture.productA.ID).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if reloaded.Price != 4000 {
		t.Fatalf("want price 4000, got %d", reloaded.Price)
	}
}

func TestProductRepository_CountByMerchant(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)

	if err := db.Create(&models.Product{
		Name:       "Cheap",
		Slug:       "cheap",
		Sku:        "CHEAP-001",
		Price:      100,
		Stock:      1,
		CategoryID: fixture.category.ID,
		MerchantID: fixture.merchant.ID,
	}).Error; err != nil {
		t.Fatalf("create cheap product: %v", err)
	}

	repo := productRepo(db)

	count, err := repo.CountByMerchant(fixture.merchant.ID)
	if err != nil {
		t.Fatalf("CountByMerchant: %v", err)
	}
	if count != 3 {
		t.Fatalf("want 3 products, got %d", count)
	}

	low, err := repo.CountLowStockByMerchant(fixture.merchant.ID, 3)
	if err != nil {
		t.Fatalf("CountLowStockByMerchant: %v", err)
	}
	if low != 2 {
		t.Fatalf("want 2 low stock (stock <= 3), got %d", low)
	}
}
