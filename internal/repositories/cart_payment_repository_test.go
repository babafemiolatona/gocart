package repositories

import (
	"errors"
	"testing"

	"gocart/internal/models"

	"gorm.io/gorm"
)

func TestPaymentRepository_TransitionStatus(t *testing.T) {
	db := openTestDB(t)
	repo := paymentRepo(db)

	payment := &models.Payment{
		Reference: "pay-1",
		Amount:    6000,
		Status:    models.PaymentStatusPending,
	}
	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		ok, err := repo.TransitionStatus("pay-1", models.PaymentStatusPending, models.PaymentStatusSucceeded)
		if err != nil {
			t.Fatalf("TransitionStatus: %v", err)
		}
		if !ok {
			t.Fatal("want transition to succeed")
		}
	})

	t.Run("wrong from status rejected", func(t *testing.T) {
		ok, err := repo.TransitionStatus("pay-1", models.PaymentStatusPending, models.PaymentStatusFailed)
		if err != nil {
			t.Fatalf("TransitionStatus: %v", err)
		}
		if ok {
			t.Fatal("want transition to be rejected")
		}
	})

	t.Run("missing reference", func(t *testing.T) {
		ok, err := repo.TransitionStatus("pay-missing", models.PaymentStatusPending, models.PaymentStatusFailed)
		if err != nil {
			t.Fatalf("TransitionStatus: %v", err)
		}
		if ok {
			t.Fatal("want no rows affected")
		}
	})
}

func TestCartRepository_ItemOps(t *testing.T) {
	db := openTestDB(t)
	users := seedUsers(t, db, 1)
	repo := cartRepo(db)

	cart := &models.Cart{UserID: users[0].ID}
	if err := db.Create(cart).Error; err != nil {
		t.Fatalf("create cart: %v", err)
	}

	item := &models.CartItem{
		CartID:    cart.ID,
		ProductID: 1,
		Quantity:  2,
		Price:     3000,
	}
	if err := repo.AddItem(item); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	t.Run("UpdateItem", func(t *testing.T) {
		item.Quantity = 5
		item.Price = 2500
		if err := repo.UpdateItem(item); err != nil {
			t.Fatalf("UpdateItem: %v", err)
		}

		var reloaded models.CartItem
		if err := db.First(&reloaded, item.ID).Error; err != nil {
			t.Fatalf("reload item: %v", err)
		}
		if reloaded.Quantity != 5 || reloaded.Price != 2500 {
			t.Fatalf("item not updated: qty=%d price=%d", reloaded.Quantity, reloaded.Price)
		}
	})

	t.Run("UpdateCartTotal", func(t *testing.T) {
		if err := repo.UpdateCartTotal(cart.ID, 12500, 5); err != nil {
			t.Fatalf("UpdateCartTotal: %v", err)
		}
		var reloaded models.Cart
		if err := db.First(&reloaded, cart.ID).Error; err != nil {
			t.Fatalf("reload cart: %v", err)
		}
		if reloaded.Total != 12500 || reloaded.ItemCount != 5 {
			t.Fatalf("cart totals not updated: total=%d count=%d", reloaded.Total, reloaded.ItemCount)
		}
	})

	t.Run("RemoveItem", func(t *testing.T) {
		if err := repo.RemoveItem(item.ID); err != nil {
			t.Fatalf("RemoveItem: %v", err)
		}
		if err := repo.RemoveItem(item.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("want ErrRecordNotFound, got %v", err)
		}
	})
}

func TestCartRepository_ClearCart(t *testing.T) {
	db := openTestDB(t)
	users := seedUsers(t, db, 1)
	repo := cartRepo(db)

	cart := &models.Cart{UserID: users[0].ID}
	if err := db.Create(cart).Error; err != nil {
		t.Fatalf("create cart: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := repo.AddItem(&models.CartItem{
			CartID:    cart.ID,
			ProductID: uint(i + 1),
			Quantity:  1,
			Price:     100,
		}); err != nil {
			t.Fatalf("AddItem %d: %v", i, err)
		}
	}

	if err := repo.ClearCart(cart.ID); err != nil {
		t.Fatalf("ClearCart: %v", err)
	}

	var itemCount int64
	if err := db.Model(&models.CartItem{}).Where("cart_id = ?", cart.ID).Count(&itemCount).Error; err != nil {
		t.Fatalf("count items: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("want 0 items after clear, got %d", itemCount)
	}

	var reloaded models.Cart
	if err := db.First(&reloaded, cart.ID).Error; err != nil {
		t.Fatalf("reload cart: %v", err)
	}
	if reloaded.Total != 0 || reloaded.ItemCount != 0 {
		t.Fatalf("cart totals not zeroed: total=%d count=%d", reloaded.Total, reloaded.ItemCount)
	}
}

func TestCartRepository_GetWithItems_Preloads(t *testing.T) {
	db := openTestDB(t)
	fixture := seedProductFixture(t, db)
	users := seedUsers(t, db, 1)
	repo := cartRepo(db)

	cart := &models.Cart{UserID: users[0].ID}
	if err := db.Create(cart).Error; err != nil {
		t.Fatalf("create cart: %v", err)
	}
	if err := repo.AddItem(&models.CartItem{
		CartID:    cart.ID,
		ProductID: fixture.productA.ID,
		Quantity:  1,
		Price:     fixture.productA.Price,
	}); err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	got, err := repo.GetWithItems(users[0].ID)
	if err != nil {
		t.Fatalf("GetWithItems: %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("want 1 item, got %d", len(got.Items))
	}
	if got.Items[0].Product.ID == 0 {
		t.Fatal("want product preloaded")
	}
}