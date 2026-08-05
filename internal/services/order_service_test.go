package services

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

func newTestOrderService(
	scope *fakeScope,
	orderRepo *stubOrderRepo,
	cartRepo *stubCartRepo,
	productRepo *stubProductRepo,
	paymentRepo *stubPaymentRepo,
) *OrderService {
	if scope == nil {
		scope = &fakeScope{
			order:   orderRepo,
			cart:    cartRepo,
			product: productRepo,
			payment: paymentRepo,
		}
	}
	return NewOrderService(&fakeTxManager{scope: scope}, orderRepo, cartRepo, productRepo, paymentRepo)
}

func testCartWithItems() *models.Cart {
	return &models.Cart{
		ID:    1,
		Total: 1999,
		Items: []models.CartItem{
			{
				ID:        1,
				ProductID: 3,
				Quantity:  1,
				Price:     1999,
				Product:   models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999},
			},
		},
	}
}

func TestCheckoutSuccess(t *testing.T) {
	var createdOrder *models.Order
	var createdPayment *models.Payment

	orderRepo := &stubOrderRepo{
		createFn: func(o *models.Order) error {
			createdOrder = o
			o.ID = 1
			return nil
		},
	}
	paymentRepo := &stubPaymentRepo{
		createFn: func(p *models.Payment) error { createdPayment = p; return nil },
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999}, nil
		},
	}

	svc := newTestOrderService(nil, orderRepo, cartRepo, productRepo, paymentRepo)

	resp, err := svc.ProcessCheckout(1, "1 Main St", "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createdOrder == nil || createdOrder.UserID != 1 || createdOrder.Status != models.OrderStatusPendingPayment {
		t.Errorf("order was not created correctly: %+v", createdOrder)
	}
	if createdOrder.ShippingAddress != "1 Main St" || createdOrder.IdempotencyKey != "key-123" {
		t.Errorf("order fields not set: %+v", createdOrder)
	}
	if len(createdOrder.Items) != 1 || createdOrder.Items[0].ProductID != 3 {
		t.Errorf("order items not copied: %+v", createdOrder.Items)
	}

	if createdPayment == nil {
		t.Fatal("expected payment to be created")
	}
	if createdPayment.OrderID != 1 || createdPayment.Reference == "" {
		t.Errorf("payment not linked to order: %+v", createdPayment)
	}
	if createdPayment.Status != models.PaymentStatusPending || createdPayment.Amount != 1999 {
		t.Errorf("payment fields wrong: %+v", createdPayment)
	}

	if resp == nil || resp.Order == nil || resp.Payment == nil {
		t.Fatal("expected checkout response")
	}
	if resp.Payment.Reference == "" {
		t.Error("expected a payment reference in the response")
	}
}

func TestCheckoutEmptyCart(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 1, Items: []models.CartItem{}}, nil
		},
	}
	svc := newTestOrderService(nil, &stubOrderRepo{}, cartRepo, &stubProductRepo{}, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "")
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeCartEmpty)
}

func TestCheckoutInsufficientStock(t *testing.T) {
	cart := testCartWithItems()
	cart.Items[0].Quantity = 5
	cart.Total = 5 * 1999

	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return cart, nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 2, Price: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, &stubOrderRepo{}, cartRepo, productRepo, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if appErr, ok := err.(*apperrors.AppError); ok {
		if appErr.Code != "insufficient_stock" {
			t.Errorf("expected insufficient_stock, got %q", appErr.Code)
		}
	} else {
		t.Errorf("expected AppError, got %T", err)
	}
}

func TestCheckoutIdempotentReplayReturnsExisting(t *testing.T) {
	existing := &models.Order{ID: 5, UserID: 1, Status: models.OrderStatusPendingPayment}
	payment := &models.Payment{ID: 9, OrderID: 5, Reference: "PAY_EXISTING", Status: models.PaymentStatusPending}

	orderRepo := &stubOrderRepo{
		getByIdemFn: func(userID uint, key string) (*models.Order, error) {
			return existing, nil
		},
	}
	paymentRepo := &stubPaymentRepo{
		getByOrderIDFn: func(orderID uint) (*models.Payment, error) { return payment, nil },
	}
	cartRepo := &stubCartRepo{}
	svc := newTestOrderService(nil, orderRepo, cartRepo, &stubProductRepo{}, paymentRepo)

	resp, err := svc.ProcessCheckout(1, "1 Main St", "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Order.ID != 5 || resp.Payment.Reference != "PAY_EXISTING" {
		t.Errorf("expected existing order/payment back, got %+v", resp)
	}
}

func TestCheckoutDuplicateKeyRecovers(t *testing.T) {
	// Race: a concurrent checkout created the order before our tx commits.
	orderRepo := &stubOrderRepo{
		createFn: func(o *models.Order) error { return repositories.ErrDuplicate },
		getByIdemFn: func(userID uint, key string) (*models.Order, error) {
			return &models.Order{ID: 7, UserID: 1, Status: models.OrderStatusPendingPayment}, nil
		},
	}
	paymentRepo := &stubPaymentRepo{
		getByOrderIDFn: func(orderID uint) (*models.Payment, error) {
			return &models.Payment{ID: 2, OrderID: 7, Reference: "PAY_RECOVER", Status: models.PaymentStatusPending}, nil
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, cartRepo, productRepo, paymentRepo)

	resp, err := svc.ProcessCheckout(1, "1 Main St", "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Order.ID != 7 || resp.Payment.Reference != "PAY_RECOVER" {
		t.Errorf("expected recovered order/payment, got %+v", resp)
	}
}

func TestCancelOrderNotOwner(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 2, Status: models.OrderStatusPendingPayment}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	err := svc.CancelOrder(1, 1)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeOrderNotFound)
}

func TestCancelOrderAlreadyCancelled(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 1, Status: models.OrderStatusCancelled}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	err := svc.CancelOrder(1, 1)
	assertAppError(t, err, http.StatusConflict, apperrors.CodeOrderAlreadyClosed)
}

func TestCancelOrderPendingPayment(t *testing.T) {
	var updatedStatus models.OrderStatus
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 1, Status: models.OrderStatusPendingPayment}, nil
		},
		updateStatusFn: func(orderID uint, status models.OrderStatus) error {
			updatedStatus = status
			return nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	if err := svc.CancelOrder(1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedStatus != models.OrderStatusCancelled {
		t.Errorf("expected status cancelled, got %q", updatedStatus)
	}
}

func TestCancelOrderConfirmedRestoresStock(t *testing.T) {
	var incremented [][2]uint

	scope := &fakeScope{
		order: &stubOrderRepo{
			transitionFn: func(orderID uint, from, to models.OrderStatus) (bool, error) {
				if from != models.OrderStatusConfirmed || to != models.OrderStatusCancelled {
					t.Errorf("unexpected transition %q -> %q", from, to)
				}
				return true, nil
			},
		},
		product: &stubProductRepo{
			incrementFn: func(id uint, qty int) error {
				incremented = append(incremented, [2]uint{id, uint(qty)})
				return nil
			},
		},
	}

	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{
				ID:     1,
				UserID: 1,
				Status: models.OrderStatusConfirmed,
				Items: []models.OrderItem{
					{ProductID: 3, Quantity: 2},
					{ProductID: 4, Quantity: 1},
				},
			}, nil
		},
	}
	svc := newTestOrderService(scope, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	if err := svc.CancelOrder(1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := [][2]uint{{3, 2}, {4, 1}}
	if len(incremented) != len(want) {
		t.Fatalf("expected %d stock increments, got %d", len(want), len(incremented))
	}
	for i := range want {
		if incremented[i] != want[i] {
			t.Errorf("increment %d = %v, want %v", i, incremented[i], want[i])
		}
	}
}

func TestCancelOrderConfirmedNotClaimedDoesNothing(t *testing.T) {
	var incrementCalls int
	scope := &fakeScope{
		order: &stubOrderRepo{
			transitionFn: func(orderID uint, from, to models.OrderStatus) (bool, error) {
				return false, nil
			},
		},
		product: &stubProductRepo{
			incrementFn: func(id uint, qty int) error { incrementCalls++; return nil },
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 1, Status: models.OrderStatusConfirmed}, nil
		},
	}
	svc := newTestOrderService(scope, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	if err := svc.CancelOrder(1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if incrementCalls != 0 {
		t.Errorf("expected no stock increments, got %d", incrementCalls)
	}
}

func TestCancelOrderInvalidStatus(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 1, Status: models.OrderStatusShipped}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	err := svc.CancelOrder(1, 1)
	assertAppError(t, err, http.StatusConflict, apperrors.CodeInvalidOrderStatus)
}

func TestGetOrderNotOwned(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 2}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	_, err := svc.GetOrder(1, 1)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeOrderNotFound)
}

func TestGetUserOrderSuccess(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 1, UserID: 1, Status: models.OrderStatusPending, Total: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	resp, err := svc.GetOrder(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.ID != 1 || resp.Status != "pending" || resp.Total != 19.99 {
		t.Errorf("unexpected order response: %+v", resp)
	}
}

func TestGetUserOrders(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByUserFn: func(userID uint, p *dto.PaginationQuery) ([]models.Order, int64, error) {
			return []models.Order{{ID: 1, Status: models.OrderStatusPending, Total: 1000}}, 1, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	resp, err := svc.GetUserOrders(1, &dto.PaginationQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 || resp.TotalPage != 1 || len(resp.Data.([]dto.OrderResponse)) != 1 {
		t.Errorf("unexpected paginated response: %+v", resp)
	}
}

func TestCheckoutValidateCartProductNotFound(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) { return nil, repositories.ErrRecordNotFound },
	}
	svc := newTestOrderService(nil, &stubOrderRepo{}, cartRepo, productRepo, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "")
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeProductNotFound)
}

func TestCheckoutValidateCartFetchFails(t *testing.T) {
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) { return nil, errBoom },
	}
	svc := newTestOrderService(nil, &stubOrderRepo{}, cartRepo, productRepo, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "")
	assertAppError(t, err, http.StatusInternalServerError, "fetch_product_failed")
}

func TestCheckoutIdempotencyLookupFails(t *testing.T) {
	orderRepo := &stubOrderRepo{
		getByIdemFn: func(userID uint, key string) (*models.Order, error) { return nil, errBoom },
	}
	svc := newTestOrderService(nil, orderRepo, &stubCartRepo{}, &stubProductRepo{}, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "key-123")
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchOrder)
}

func TestCheckoutOrderCreateFails(t *testing.T) {
	orderRepo := &stubOrderRepo{
		createFn: func(o *models.Order) error { return errBoom },
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, cartRepo, productRepo, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "")
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeCreateOrder)
}

func TestCheckoutPaymentCreateFails(t *testing.T) {
	orderRepo := &stubOrderRepo{
		createFn: func(o *models.Order) error { o.ID = 1; return nil },
	}
	paymentRepo := &stubPaymentRepo{
		createFn: func(p *models.Payment) error { return errBoom },
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, cartRepo, productRepo, paymentRepo)

	_, err := svc.ProcessCheckout(1, "1 Main St", "")
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeCreatePayment)
}

func TestCheckoutConflictRecoveryOrderFetchFails(t *testing.T) {
	orderRepo := &stubOrderRepo{
		createFn: func(o *models.Order) error { return repositories.ErrDuplicate },
		getByIdemFn: func(userID uint, key string) (*models.Order, error) {
			return nil, errBoom
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, cartRepo, productRepo, &stubPaymentRepo{})

	_, err := svc.ProcessCheckout(1, "1 Main St", "key-123")
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchOrder)
}

func TestCheckoutConflictRecoveryPaymentFetchFails(t *testing.T) {
	orderRepo := &stubOrderRepo{
		createFn: func(o *models.Order) error { return repositories.ErrDuplicate },
		getByIdemFn: func(userID uint, key string) (*models.Order, error) {
			return &models.Order{ID: 7, UserID: 1}, nil
		},
	}
	paymentRepo := &stubPaymentRepo{
		getByOrderIDFn: func(orderID uint) (*models.Payment, error) { return nil, errBoom },
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) { return testCartWithItems(), nil },
	}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 3, MerchantID: 1, Stock: 10, Price: 1999}, nil
		},
	}
	svc := newTestOrderService(nil, orderRepo, cartRepo, productRepo, paymentRepo)

	_, err := svc.ProcessCheckout(1, "1 Main St", "key-123")
	assertAppError(t, err, http.StatusInternalServerError, "fetch_payment_failed")
}
