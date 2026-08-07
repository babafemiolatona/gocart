package repositories

import (
	"fmt"
	"sync/atomic"
	"testing"

	"gocart/internal/models"

	sqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDBSeq int64

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", atomic.AddInt64(&testDBSeq, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Merchant{},
		&models.Category{},
		&models.Product{},
		&models.ProductImage{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

type seedProducts struct {
	merchant *models.Merchant
	category *models.Category
	productA *models.Product
	productB *models.Product
}

// seedProductFixture creates the category, merchant, and two products used by
// the order and product repository tests.
func seedProductFixture(t *testing.T, db *gorm.DB) *seedProducts {
	t.Helper()

	category := &models.Category{Name: "Books", Slug: "books"}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	merchant := &models.Merchant{UserID: 1, BusinessName: "Acme"}
	if err := db.Create(merchant).Error; err != nil {
		t.Fatalf("create merchant: %v", err)
	}

	productA := &models.Product{
		Name:       "The Go Programming Language",
		Slug:       "go-book",
		Sku:        "GO-001",
		Price:      3000,
		Stock:      10,
		CategoryID: category.ID,
		MerchantID: merchant.ID,
	}
	productB := &models.Product{
		Name:       "Concurrency in Go",
		Slug:       "go-concur",
		Sku:        "GO-002",
		Price:      5000,
		Stock:      0,
		CategoryID: category.ID,
		MerchantID: merchant.ID,
	}
	if err := db.Create(&productA).Error; err != nil {
		t.Fatalf("create product A: %v", err)
	}
	if err := db.Create(&productB).Error; err != nil {
		t.Fatalf("create product B: %v", err)
	}

	return &seedProducts{
		merchant: merchant,
		category: category,
		productA: productA,
		productB: productB,
	}
}

// createOrderForUser seeds a user and an order (with its items) and returns them.
func createOrderForUser(t *testing.T, db *gorm.DB, userID uint, fixture *seedProducts, total int64) *models.Order {
	t.Helper()

	order := &models.Order{
		UserID:          userID,
		Status:          models.OrderStatusPending,
		Total:           total,
		ShippingAddress: "123 Main St",
		Items: []models.OrderItem{
			{
				ProductID:   fixture.productA.ID,
				ProductName: fixture.productA.Name,
				Quantity:    2,
				Price:       fixture.productA.Price,
			},
		},
	}

	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	return order
}

func orderRepo(db *gorm.DB) OrderRepository { return NewOrderRepository(db) }
func productRepo(db *gorm.DB) ProductRepository {
	return NewProductRepository(db)
}
func paymentRepo(db *gorm.DB) PaymentRepository { return NewPaymentRepository(db) }
func cartRepo(db *gorm.DB) CartRepository       { return NewCartRepository(db) }

func mustUser(t *testing.T, db *gorm.DB, id uint) *models.User {
	t.Helper()
	user := &models.User{}
	if err := db.First(user, id).Error; err != nil {
		t.Fatalf("load user %d: %v", id, err)
	}
	return user
}

func seedUsers(t *testing.T, db *gorm.DB, n int) []*models.User {
	t.Helper()
	users := make([]*models.User, 0, n)
	for i := 0; i < n; i++ {
		u := &models.User{
			Username: fmt.Sprintf("user%d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Role:     models.RoleCustomer,
		}
		if err := db.Create(u).Error; err != nil {
			t.Fatalf("create user %d: %v", i+1, err)
		}
		users = append(users, u)
	}
	return users
}
