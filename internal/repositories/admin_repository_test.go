package repositories

import (
	"testing"

	"gocart/internal/models"
)

func TestAdminDashboardAggregates(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 2)
	repo := orderRepo(db)

	delivered := createOrderForUser(t, db, users[0].ID, fixture, 6000)
	delivered.Status = models.OrderStatusDelivered
	if err := db.Save(delivered).Error; err != nil {
		t.Fatalf("mark delivered: %v", err)
	}

	createOrderForUser(t, db, users[1].ID, fixture, 3000)

	cancelled := createOrderForUser(t, db, users[0].ID, fixture, 9999)
	cancelled.Status = models.OrderStatusCancelled
	if err := db.Save(cancelled).Error; err != nil {
		t.Fatalf("mark cancelled: %v", err)
	}

	otherMerchant := &models.Merchant{UserID: 99, BusinessName: "Other Shop"}
	if err := db.Create(otherMerchant).Error; err != nil {
		t.Fatalf("create merchant: %v", err)
	}

	t.Run("CountAll users", func(t *testing.T) {
		count, err := NewUserRepository(db).CountAll()
		if err != nil {
			t.Fatalf("CountAll users: %v", err)
		}
		if count != 2 {
			t.Fatalf("want 2 users, got %d", count)
		}
	})

	t.Run("CountAll merchants", func(t *testing.T) {
		count, err := NewMerchantRepository(db).CountAll()
		if err != nil {
			t.Fatalf("CountAll merchants: %v", err)
		}
		if count != 2 {
			t.Fatalf("want 2 merchants, got %d", count)
		}
	})

	t.Run("CountAll products", func(t *testing.T) {
		count, err := NewProductRepository(db).CountAll()
		if err != nil {
			t.Fatalf("CountAll products: %v", err)
		}
		if count != 2 {
			t.Fatalf("want 2 products, got %d", count)
		}
	})

	t.Run("CountAll orders", func(t *testing.T) {
		count, err := repo.CountAll()
		if err != nil {
			t.Fatalf("CountAll orders: %v", err)
		}
		if count != 3 {
			t.Fatalf("want 3 orders, got %d", count)
		}
	})

	t.Run("SumRevenueAll only delivered", func(t *testing.T) {
		total, err := repo.SumRevenueAll()
		if err != nil {
			t.Fatalf("SumRevenueAll: %v", err)
		}
		if total != 6000 {
			t.Fatalf("want revenue 6000 (delivered only), got %d", total)
		}
	})

	t.Run("CountsByStatus groups correctly", func(t *testing.T) {
		counts, err := repo.CountsByStatus()
		if err != nil {
			t.Fatalf("CountsByStatus: %v", err)
		}
		if counts[models.OrderStatusDelivered] != 1 {
			t.Fatalf("want 1 delivered, got %d", counts[models.OrderStatusDelivered])
		}
		if counts[models.OrderStatusPending] != 1 {
			t.Fatalf("want 1 pending, got %d", counts[models.OrderStatusPending])
		}
		if counts[models.OrderStatusCancelled] != 1 {
			t.Fatalf("want 1 cancelled, got %d", counts[models.OrderStatusCancelled])
		}
	})
}
