package repositories

import (
	"errors"
	"testing"

	"gocart/internal/dto"
	"gocart/internal/models"

	"gorm.io/gorm"
)

func TestOrderRepository_GetByUserIDAndIdempotencyKey(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 1)
	repo := orderRepo(db)

	order := createOrderForUser(t, db, users[0].ID, fixture, 6000)
	order.IdempotencyKey = "order-idem-1"
	if err := db.Save(order).Error; err != nil {
		t.Fatalf("set idempotency key: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetByUserIDAndIdempotencyKey(users[0].ID, "order-idem-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != order.ID {
			t.Fatalf("want order %d, got %d", order.ID, got.ID)
		}
		if len(got.Items) != 1 {
			t.Fatalf("want 1 preloaded item, got %d", len(got.Items))
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := repo.GetByUserIDAndIdempotencyKey(users[0].ID, "missing"); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("want ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("scoped to user", func(t *testing.T) {
		other := &models.User{Username: "other", Email: "other@example.com", Role: models.RoleCustomer}
		if err := db.Create(other).Error; err != nil {
			t.Fatalf("create other user: %v", err)
		}
		if _, err := repo.GetByUserIDAndIdempotencyKey(other.ID, "order-idem-1"); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("want ErrRecordNotFound for other user, got %v", err)
		}
	})
}

func TestOrderRepository_UpdateOrderStatus(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 1)
	repo := orderRepo(db)

	order := createOrderForUser(t, db, users[0].ID, fixture, 6000)

	if err := repo.UpdateOrderStatus(order.ID, models.OrderStatusConfirmed); err != nil {
		t.Fatalf("update status: %v", err)
	}

	var updated models.Order
	if err := db.First(&updated, order.ID).Error; err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if updated.Status != models.OrderStatusConfirmed {
		t.Fatalf("want status confirmed, got %s", updated.Status)
	}

	if err := repo.UpdateOrderStatus(99999, models.OrderStatusConfirmed); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("want ErrRecordNotFound, got %v", err)
	}
}

func TestOrderRepository_TransitionOrderStatus(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 1)
	repo := orderRepo(db)

	order := createOrderForUser(t, db, users[0].ID, fixture, 6000)

	t.Run("success", func(t *testing.T) {
		ok, err := repo.TransitionOrderStatus(order.ID, models.OrderStatusPending, models.OrderStatusConfirmed)
		if err != nil {
			t.Fatalf("transition: %v", err)
		}
		if !ok {
			t.Fatal("want transition to succeed")
		}
	})

	t.Run("wrong from status is a no-op", func(t *testing.T) {
		ok, err := repo.TransitionOrderStatus(order.ID, models.OrderStatusPending, models.OrderStatusShipped)
		if err != nil {
			t.Fatalf("transition: %v", err)
		}
		if ok {
			t.Fatal("want transition to be rejected")
		}

		var current models.Order
		if err := db.First(&current, order.ID).Error; err != nil {
			t.Fatalf("reload order: %v", err)
		}
		if current.Status != models.OrderStatusConfirmed {
			t.Fatalf("want status to stay confirmed, got %s", current.Status)
		}
	})

	t.Run("missing order", func(t *testing.T) {
		ok, err := repo.TransitionOrderStatus(99999, models.OrderStatusPending, models.OrderStatusConfirmed)
		if err != nil {
			t.Fatalf("transition: %v", err)
		}
		if ok {
			t.Fatal("want no rows affected")
		}
	})
}

func TestOrderRepository_GetOrdersByUserID_Pagination(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 2)
	repo := orderRepo(db)

	user := users[0]
	for i := 0; i < 3; i++ {
		createOrderForUser(t, db, user.ID, fixture, int64(1000+i))
	}
	createOrderForUser(t, db, users[1].ID, fixture, 9999)

	orders, total, err := repo.GetOrdersByUserID(user.ID, &dto.PaginationQuery{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if total != 3 {
		t.Fatalf("want total 3, got %d", total)
	}
	if len(orders) != 2 {
		t.Fatalf("want 2 orders on page 1, got %d", len(orders))
	}

	page2, _, err := repo.GetOrdersByUserID(user.ID, &dto.PaginationQuery{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("query page 2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("want 1 order on page 2, got %d", len(page2))
	}

	asc, _, err := repo.GetOrdersByUserID(user.ID, &dto.PaginationQuery{Page: 1, PageSize: 10, Order: "asc"})
	if err != nil {
		t.Fatalf("query asc: %v", err)
	}
	if len(asc) != 3 {
		t.Fatalf("want 3 orders asc, got %d", len(asc))
	}
	if asc[0].Total != 1000 {
		t.Fatalf("want smallest total first in asc, got %d", asc[0].Total)
	}
}

func TestOrderRepository_MerchantDashboardMetrics(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 1)
	repo := orderRepo(db)

	delivered := createOrderForUser(t, db, users[0].ID, fixture, 6000)
	delivered.Status = models.OrderStatusDelivered
	if err := db.Save(delivered).Error; err != nil {
		t.Fatalf("mark delivered: %v", err)
	}

	otherDelivered := createOrderForUser(t, db, users[0].ID, fixture, 2000)
	otherDelivered.Status = models.OrderStatusDelivered
	if err := db.Save(otherDelivered).Error; err != nil {
		t.Fatalf("mark delivered: %v", err)
	}

	cancelled := createOrderForUser(t, db, users[0].ID, fixture, 9999)
	cancelled.Status = models.OrderStatusCancelled
	if err := db.Save(cancelled).Error; err != nil {
		t.Fatalf("mark cancelled: %v", err)
	}

	count, err := repo.CountByMerchant(fixture.merchant.ID)
	if err != nil {
		t.Fatalf("CountByMerchant: %v", err)
	}
	if count != 3 {
		t.Fatalf("want 3 orders, got %d", count)
	}

	deliveredCount, err := repo.CountByMerchantAndStatus(fixture.merchant.ID, models.OrderStatusDelivered)
	if err != nil {
		t.Fatalf("CountByMerchantAndStatus: %v", err)
	}
	if deliveredCount != 2 {
		t.Fatalf("want 2 delivered, got %d", deliveredCount)
	}

	revenue, err := repo.SumRevenueByMerchant(fixture.merchant.ID)
	if err != nil {
		t.Fatalf("SumRevenueByMerchant: %v", err)
	}
	if revenue != 8000 {
		t.Fatalf("want revenue 8000 (delivered only), got %d", revenue)
	}

}
