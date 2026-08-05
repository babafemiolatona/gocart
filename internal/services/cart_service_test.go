package services

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

func newTestCartService(cartRepo *stubCartRepo, productRepo *stubProductRepo) *CartService {
	return NewCartService(cartRepo, productRepo)
}

func TestGetCartReturnsExisting(t *testing.T) {
	cart := &models.Cart{ID: 1, UserID: 7}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return cart, nil },
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	got, err := svc.GetCart(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cart {
		t.Errorf("expected the configured cart back")
	}
}

func TestGetCartCreatesWhenMissing(t *testing.T) {
	var created *models.Cart
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return nil, repositories.ErrRecordNotFound
		},
		createFn: func(c *models.Cart) error { created = c; return nil },
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	got, err := svc.GetCart(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil || created.UserID != 7 {
		t.Errorf("expected a new cart to be created for user 7")
	}
	if got.UserID != 7 {
		t.Errorf("expected returned cart for user 7")
	}
}

func TestGetCartFailsOnRepoError(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return nil, errBoom },
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	_, err := svc.GetCart(7)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchCart)
}

func TestGetCartResponse(t *testing.T) {
	cart := &models.Cart{
		ID:        1,
		UserID:    7,
		Total:     2000,
		ItemCount: 2,
		Items: []models.CartItem{{
			ProductID: 3,
			Quantity:  2,
			Price:     1000,
			Product:   models.Product{ID: 3, Name: "Widget", Slug: "widget"},
		}},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return cart, nil },
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	resp, err := svc.GetCartResponse(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ItemCount != 2 || resp.Total != 20.0 {
		t.Errorf("unexpected cart response: %+v", resp)
	}
	if len(resp.Items) != 1 || resp.Items[0].Product.Name != "Widget" {
		t.Errorf("unexpected cart items: %+v", resp.Items)
	}
}

func TestAddToCartInvalidQuantity(t *testing.T) {
	svc := newTestCartService(&stubCartRepo{}, &stubProductRepo{})

	_, err := svc.AddToCart(1, &dto.AddToCartRequest{ProductID: 1, Quantity: 0})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidQuantity)
}

func TestAddToCartProductNotFound(t *testing.T) {
	productRepo := &stubProductRepo{}
	svc := newTestCartService(&stubCartRepo{}, productRepo)

	_, err := svc.AddToCart(1, &dto.AddToCartRequest{ProductID: 99, Quantity: 1})
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeProductNotFound)
}

func TestAddToCartDifferentMerchantConflict(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{
				ID: 1,
				Items: []models.CartItem{
					{ID: 1, Product: models.Product{MerchantID: 1}},
				},
			}, nil
		},
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 2, MerchantID: 2, Stock: 10}, nil
		},
	}
	svc := newTestCartService(cartRepo, productRepo)

	_, err := svc.AddToCart(1, &dto.AddToCartRequest{ProductID: 2, Quantity: 1})
	assertAppError(t, err, http.StatusConflict, apperrors.CodeMultipleMerchants)
}

func TestAddToCartNewItem(t *testing.T) {
	var added *models.CartItem
	items := []models.CartItem{}
	var totalCall int64

	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 1, Items: items}, nil
		},
		addItemFn: func(i *models.CartItem) error {
			added = i
			items = append(items, *i)
			return nil
		},
		updateTotalFn: func(cartID uint, total int64, count int) error {
			totalCall = total
			return nil
		},
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, Name: "Widget", MerchantID: 1, Price: 1999, Stock: 10}, nil
		},
	}
	svc := newTestCartService(cartRepo, productRepo)

	resp, err := svc.AddToCart(1, &dto.AddToCartRequest{ProductID: 3, Quantity: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if added == nil || added.ProductID != 3 || added.Quantity != 2 || added.Price != 1999 {
		t.Errorf("item was not added correctly: %+v", added)
	}
	if totalCall != 3998 {
		t.Errorf("expected recalculated total 3998, got %d", totalCall)
	}
	if resp == nil {
		t.Error("expected a response")
	}
}

func TestAddToCartMergesExistingItem(t *testing.T) {
	var updated *models.CartItem

	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{
				ID: 1,
				Items: []models.CartItem{
					{ID: 10, ProductID: 3, Quantity: 1, Price: 1999, Product: models.Product{MerchantID: 1}},
				},
			}, nil
		},
		updateItemFn: func(i *models.CartItem) error { updated = i; return nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Price: 1999, Stock: 100}, nil
		},
	}
	svc := newTestCartService(cartRepo, productRepo)

	_, err := svc.AddToCart(1, &dto.AddToCartRequest{ProductID: 3, Quantity: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated == nil || updated.Quantity != 4 {
		t.Errorf("expected merged quantity 4, got %+v", updated)
	}
}

func TestAddToCartInsufficientStock(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 1}, nil
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 1, Items: []models.CartItem{}}, nil
		},
	}
	svc := newTestCartService(cartRepo, productRepo)

	_, err := svc.AddToCart(1, &dto.AddToCartRequest{ProductID: 3, Quantity: 5})
	assertAppError(t, err, http.StatusConflict, apperrors.CodeInsufficientStock)
}

func TestUpdateCartItemInvalidQuantity(t *testing.T) {
	svc := newTestCartService(&stubCartRepo{}, &stubProductRepo{})

	_, err := svc.UpdateCartItem(1, 1, 0)
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidQuantity)
}

func TestUpdateCartItemNotFound(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 1, Items: []models.CartItem{}}, nil
		},
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	_, err := svc.UpdateCartItem(1, 42, 2)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeCartItemNotFound)
}

func TestUpdateCartItemSuccess(t *testing.T) {
	var updated *models.CartItem

	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{
				ID: 1,
				Items: []models.CartItem{
					{ID: 10, ProductID: 3, Quantity: 1},
				},
			}, nil
		},
		updateItemFn: func(i *models.CartItem) error { updated = i; return nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Price: 1999, Stock: 50}, nil
		},
	}
	svc := newTestCartService(cartRepo, productRepo)

	_, err := svc.UpdateCartItem(1, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated == nil || updated.Quantity != 5 {
		t.Errorf("expected quantity 5, got %+v", updated)
	}
}

func TestUpdateCartItemInsufficientStock(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{
				ID: 1,
				Items: []models.CartItem{
					{ID: 10, ProductID: 3, Quantity: 1},
				},
			}, nil
		},
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 2}, nil
		},
	}
	svc := newTestCartService(cartRepo, productRepo)

	_, err := svc.UpdateCartItem(1, 10, 10)
	assertAppError(t, err, http.StatusConflict, apperrors.CodeInsufficientStock)
}

func TestRemoveFromCartNotFound(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 1, Items: []models.CartItem{}}, nil
		},
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	_, err := svc.RemoveFromCart(1, 42)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeCartItemNotFound)
}

func TestRemoveFromCartSuccess(t *testing.T) {
	var removed uint

	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{
				ID: 1,
				Items: []models.CartItem{
					{ID: 10, ProductID: 3, Quantity: 1},
				},
			}, nil
		},
		removeItemFn: func(id uint) error { removed = id; return nil },
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	_, err := svc.RemoveFromCart(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 10 {
		t.Errorf("expected to remove item 10, got %d", removed)
	}
}

func TestClearCart(t *testing.T) {
	var cleared uint
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 5, UserID: 1}, nil
		},
		clearCartFn: func(cartID uint) error { cleared = cartID; return nil },
	}
	svc := newTestCartService(cartRepo, &stubProductRepo{})

	if err := svc.ClearCart(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleared != 5 {
		t.Errorf("expected to clear cart 5, got %d", cleared)
	}
}
