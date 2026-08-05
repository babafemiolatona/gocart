package services

import (
	"net/http"
	"testing"

	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

func newTestPaymentService(
	scope *fakeScope,
	paymentRepo *stubPaymentRepo,
	orderRepo *stubOrderRepo,
	cartRepo *stubCartRepo,
	productRepo *stubProductRepo,
) *PaymentService {
	if scope == nil {
		scope = &fakeScope{
			payment: paymentRepo,
			order:   orderRepo,
			cart:    cartRepo,
			product: productRepo,
		}
	}
	return NewPaymentService(&fakeTxManager{scope: scope}, paymentRepo, orderRepo, cartRepo, productRepo)
}

func pendingPaymentFixture() *models.Payment {
	return &models.Payment{
		ID:        1,
		OrderID:   10,
		Reference: "PAY_1",
		Amount:    1999,
		Status:    models.PaymentStatusPending,
		Provider:  models.PaymentProviderMock,
	}
}

func TestProcessPaymentSuccess(t *testing.T) {
	var decremented [][2]uint
	var clearedCart uint
	var updatedStatus models.OrderStatus
	var transitionCalled bool

	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
				transitionCalled = true
				if from != models.PaymentStatusPending || to != models.PaymentStatusSucceeded {
					t.Errorf("unexpected transition %q -> %q", from, to)
				}
				return true, nil
			},
		},
		product: &stubProductRepo{
			decrementFn: func(id uint, qty int) error {
				decremented = append(decremented, [2]uint{id, uint(qty)})
				return nil
			},
		},
		cart: &stubCartRepo{
			clearCartFn: func(cartID uint) error { clearedCart = cartID; return nil },
		},
		order: &stubOrderRepo{
			updateStatusFn: func(orderID uint, status models.OrderStatus) error {
				updatedStatus = status
				return nil
			},
		},
	}

	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) {
			return pendingPaymentFixture(), nil
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 1, Status: models.OrderStatusPendingPayment}, nil
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{
				ID: 5,
				Items: []models.CartItem{
					{ProductID: 3, Quantity: 2},
				},
			}, nil
		},
	}
	productRepo := &stubProductRepo{}

	svc := newTestPaymentService(scope, paymentRepo, orderRepo, cartRepo, productRepo)

	resp, err := svc.ProcessPayment(1, "PAY_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !transitionCalled {
		t.Error("expected payment transition")
	}
	if len(decremented) != 1 || decremented[0] != [2]uint{3, 2} {
		t.Errorf("expected stock decrement for product 3 qty 2, got %v", decremented)
	}
	if clearedCart != 5 {
		t.Errorf("expected cart 5 to be cleared, got %d", clearedCart)
	}
	if updatedStatus != models.OrderStatusConfirmed {
		t.Errorf("expected order confirmed, got %q", updatedStatus)
	}
	if resp == nil || resp.Status != "pending" {
		t.Errorf("expected payment response, got %+v", resp)
	}
}

func TestProcessPaymentIdempotentWhenAlreadySucceeded(t *testing.T) {
	payment := pendingPaymentFixture()
	payment.Status = models.PaymentStatusSucceeded

	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) { return payment, nil },
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 1}, nil
		},
	}
	cartRepo := &stubCartRepo{}
	svc := newTestPaymentService(nil, paymentRepo, orderRepo, cartRepo, &stubProductRepo{})

	resp, err := svc.ProcessPayment(1, "PAY_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "succeeded" {
		t.Errorf("expected succeeded payment back, got %+v", resp)
	}
}

func TestProcessPaymentNotOwner(t *testing.T) {
	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) {
			return pendingPaymentFixture(), nil
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 99}, nil
		},
	}
	svc := newTestPaymentService(nil, paymentRepo, orderRepo, &stubCartRepo{}, &stubProductRepo{})

	_, err := svc.ProcessPayment(1, "PAY_1")
	assertAppError(t, err, http.StatusNotFound, apperrors.CodePaymentNotFound)
}

func TestProcessPaymentNotFound(t *testing.T) {
	paymentRepo := &stubPaymentRepo{}
	orderRepo := &stubOrderRepo{}
	svc := newTestPaymentService(nil, paymentRepo, orderRepo, &stubCartRepo{}, &stubProductRepo{})

	_, err := svc.ProcessPayment(1, "PAY_MISSING")
	assertAppError(t, err, http.StatusNotFound, apperrors.CodePaymentNotFound)
}

func TestProcessPaymentInsufficientStock(t *testing.T) {
	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
				return true, nil
			},
		},
		product: &stubProductRepo{
			decrementFn: func(id uint, qty int) error { return repositories.ErrInsufficientStock },
		},
		cart:   &stubCartRepo{},
		order:  &stubOrderRepo{},
	}

	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) {
			return pendingPaymentFixture(), nil
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 1}, nil
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 5, Items: []models.CartItem{{ProductID: 3, Quantity: 2}}}, nil
		},
	}

	svc := newTestPaymentService(scope, paymentRepo, orderRepo, cartRepo, &stubProductRepo{})

	_, err := svc.ProcessPayment(1, "PAY_1")
	assertAppError(t, err, http.StatusConflict, apperrors.CodeInsufficientStock)
}

func TestProcessPaymentTransitionNotClaimed(t *testing.T) {
	var decrements int
	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
				return false, nil
			},
		},
		product: &stubProductRepo{
			decrementFn: func(id uint, qty int) error { decrements++; return nil },
		},
	}

	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) {
			return pendingPaymentFixture(), nil
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 1}, nil
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 5, Items: []models.CartItem{{ProductID: 3, Quantity: 2}}}, nil
		},
	}

	svc := newTestPaymentService(scope, paymentRepo, orderRepo, cartRepo, &stubProductRepo{})

	if _, err := svc.ProcessPayment(1, "PAY_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decrements != 0 {
		t.Errorf("expected no stock decrements when transition was not claimed, got %d", decrements)
	}
}

func TestGetPaymentSuccess(t *testing.T) {
	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) {
			return pendingPaymentFixture(), nil
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 1}, nil
		},
	}
	svc := newTestPaymentService(nil, paymentRepo, orderRepo, &stubCartRepo{}, &stubProductRepo{})

	resp, err := svc.GetPayment(1, "PAY_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Reference != "PAY_1" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetPaymentNotFound(t *testing.T) {
	svc := newTestPaymentService(nil, &stubPaymentRepo{}, &stubOrderRepo{}, &stubCartRepo{}, &stubProductRepo{})

	_, err := svc.GetPayment(1, "PAY_1")
	assertAppError(t, err, http.StatusNotFound, apperrors.CodePaymentNotFound)
}

func TestGetPaymentOrderMismatch(t *testing.T) {
	paymentRepo := &stubPaymentRepo{
		getByReferenceFn: func(reference string) (*models.Payment, error) {
			return pendingPaymentFixture(), nil
		},
	}
	orderRepo := &stubOrderRepo{
		getByIDFn: func(id uint) (*models.Order, error) {
			return &models.Order{ID: 10, UserID: 99}, nil
		},
	}
	svc := newTestPaymentService(nil, paymentRepo, orderRepo, &stubCartRepo{}, &stubProductRepo{})

	_, err := svc.GetPayment(1, "PAY_1")
	assertAppError(t, err, http.StatusNotFound, apperrors.CodePaymentNotFound)
}
