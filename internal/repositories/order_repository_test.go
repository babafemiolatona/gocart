package repositories

import (
	"errors"
	"testing"
	"time"

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

	multiItem := &models.Order{
		UserID:          users[0].ID,
		Status:          models.OrderStatusDelivered,
		Total:           7000,
		ShippingAddress: "321 Maple Dr",
		Items: []models.OrderItem{
			{ProductID: fixture.productA.ID, ProductName: fixture.productA.Name, Quantity: 1, Price: fixture.productA.Price},
			{ProductID: fixture.productB.ID, ProductName: fixture.productB.Name, Quantity: 1, Price: fixture.productB.Price},
		},
	}
	if err := db.Create(multiItem).Error; err != nil {
		t.Fatalf("create multi-item delivered order: %v", err)
	}

	if count, err := repo.CountByMerchant(fixture.merchant.ID); err != nil {
		t.Fatalf("CountByMerchant: %v", err)
	} else if count != 4 {
		t.Fatalf("want 4 orders (multi-item counted once), got %d", count)
	}

	if deliveredCount, err := repo.CountByMerchantAndStatus(fixture.merchant.ID, models.OrderStatusDelivered); err != nil {
		t.Fatalf("CountByMerchantAndStatus: %v", err)
	} else if deliveredCount != 3 {
		t.Fatalf("want 3 delivered, got %d", deliveredCount)
	}

	if revenue, err := repo.SumRevenueByMerchant(fixture.merchant.ID); err != nil {
		t.Fatalf("SumRevenueByMerchant: %v", err)
	} else if revenue != 15000 {
		t.Fatalf("want revenue 15000 (each delivered order once), got %d", revenue)
	}

}

func TestOrderRepository_MerchantScopedOrderQueries(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 2)
	repo := orderRepo(db)

	order1 := createOrderForUser(t, db, users[0].ID, fixture, 6000)
	order2 := createOrderForUser(t, db, users[1].ID, fixture, 3000)
	order3 := createOrderForUser(t, db, users[0].ID, fixture, 1000)

	order4 := &models.Order{
		UserID:          users[1].ID,
		Status:          models.OrderStatusPending,
		Total:           10000,
		ShippingAddress: "456 Elm St",
		Items: []models.OrderItem{
			{
				ProductID:   fixture.productA.ID,
				ProductName: fixture.productA.Name,
				Quantity:    1,
				Price:       fixture.productA.Price,
			},
			{
				ProductID:   fixture.productB.ID,
				ProductName: fixture.productB.Name,
				Quantity:    1,
				Price:       fixture.productB.Price,
			},
		},
	}
	if err := db.Create(order4).Error; err != nil {
		t.Fatalf("create multi-item order: %v", err)
	}

	base := time.Now().Add(-24 * time.Hour)
	for i, o := range []*models.Order{order1, order2, order3, order4} {
		o.CreatedAt = base.Add(time.Duration(i) * time.Hour)
		if err := db.Save(o).Error; err != nil {
			t.Fatalf("set created_at for order %d: %v", o.ID, err)
		}
	}

	otherMerchant := &models.Merchant{UserID: users[1].ID, BusinessName: "Other Books"}
	if err := db.Create(otherMerchant).Error; err != nil {
		t.Fatalf("create other merchant: %v", err)
	}
	otherProduct := &models.Product{
		Name:       "Foreign Book",
		Slug:       "foreign-book",
		Sku:        "OTH-001",
		Price:      1500,
		Stock:      5,
		CategoryID: fixture.category.ID,
		MerchantID: otherMerchant.ID,
	}
	if err := db.Create(otherProduct).Error; err != nil {
		t.Fatalf("create other product: %v", err)
	}
	otherOrder := &models.Order{
		UserID:          users[0].ID,
		Status:          models.OrderStatusPending,
		Total:           1500,
		ShippingAddress: "789 Oak Ave",
		Items: []models.OrderItem{
			{
				ProductID:   otherProduct.ID,
				ProductName: otherProduct.Name,
				Quantity:    1,
				Price:       otherProduct.Price,
			},
		},
	}
	if err := db.Create(otherOrder).Error; err != nil {
		t.Fatalf("create other merchant order: %v", err)
	}

	t.Run("GetOrdersByMerchantID dedupes and counts", func(t *testing.T) {
		orders, total, err := repo.GetOrdersByMerchantID(fixture.merchant.ID, &dto.PaginationQuery{Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		if total != 4 {
			t.Fatalf("want total 4 (deduped), got %d", total)
		}
		if len(orders) != 4 {
			t.Fatalf("want 4 orders, got %d", len(orders))
		}
	})

	t.Run("GetOrdersByMerchantID paginates newest first", func(t *testing.T) {
		page1, total, err := repo.GetOrdersByMerchantID(fixture.merchant.ID, &dto.PaginationQuery{Page: 1, PageSize: 2})
		if err != nil {
			t.Fatalf("query page 1: %v", err)
		}
		if total != 4 {
			t.Fatalf("want total 4, got %d", total)
		}
		if len(page1) != 2 {
			t.Fatalf("want 2 orders on page 1, got %d", len(page1))
		}
		if page1[0].ID != order4.ID {
			t.Fatalf("want newest order first, got %d", page1[0].ID)
		}

		page2, _, err := repo.GetOrdersByMerchantID(fixture.merchant.ID, &dto.PaginationQuery{Page: 2, PageSize: 2})
		if err != nil {
			t.Fatalf("query page 2: %v", err)
		}
		if len(page2) != 2 {
			t.Fatalf("want 2 orders on page 2, got %d", len(page2))
		}
	})

	t.Run("GetMerchantOrderByID scoped to merchant", func(t *testing.T) {
		got, err := repo.GetMerchantOrderByID(fixture.merchant.ID, order2.ID)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		if got.ID != order2.ID {
			t.Fatalf("want order %d, got %d", order2.ID, got.ID)
		}

		if _, err := repo.GetMerchantOrderByID(fixture.merchant.ID, otherOrder.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("want ErrRecordNotFound for other merchant's order, got %v", err)
		}
	})

	t.Run("GetRecentOrdersByMerchant limits newest first", func(t *testing.T) {
		recent, err := repo.GetRecentOrdersByMerchant(fixture.merchant.ID, 2)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		if len(recent) != 2 {
			t.Fatalf("want 2 recent orders, got %d", len(recent))
		}
		if recent[0].ID != order4.ID {
			t.Fatalf("want newest first, got %d", recent[0].ID)
		}
	})
}
