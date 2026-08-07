package services

import (
	"encoding/json"
	"net/http"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"gocart/internal/security"
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
	return NewPaymentService(&fakeTxManager{scope: scope}, paymentRepo, orderRepo, cartRepo, productRepo, "test-secret")
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
		cart:  &stubCartRepo{},
		order: &stubOrderRepo{},
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

func signedWebhookBody(t *testing.T, event dto.PaymentWebhookEvent) ([]byte, string) {
	t.Helper()
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	return body, security.SignWebhookBody([]byte("test-secret"), body)
}

func TestProcessWebhookInvalidSignature(t *testing.T) {
	svc := newTestPaymentService(nil, &stubPaymentRepo{}, &stubOrderRepo{}, &stubCartRepo{}, &stubProductRepo{})

	body, _ := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_1", Status: "succeeded"})

	_, err := svc.ProcessWebhook(body, "bogus-signature")
	assertAppError(t, err, http.StatusUnauthorized, apperrors.CodeInvalidWebhookSignature)

	_, err = svc.ProcessWebhook(body, "")
	assertAppError(t, err, http.StatusUnauthorized, apperrors.CodeInvalidWebhookSignature)
}

func TestProcessWebhookSucceeded(t *testing.T) {
	var decremented [][2]uint
	var clearedCart uint
	var updatedStatus models.OrderStatus

	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
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
			return &models.Order{ID: 10, UserID: 1}, nil
		},
	}
	cartRepo := &stubCartRepo{
		getWithItemsFn: func(userID uint) (*models.Cart, error) {
			return &models.Cart{ID: 5, Items: []models.CartItem{{ProductID: 3, Quantity: 2}}}, nil
		},
	}

	svc := newTestPaymentService(scope, paymentRepo, orderRepo, cartRepo, &stubProductRepo{})

	body, sig := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_1", Status: "succeeded", Amount: 1999})
	resp, err := svc.ProcessWebhook(body, sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	if resp == nil || resp.Reference != "PAY_1" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestProcessWebhookReplayIsIdempotent(t *testing.T) {
	var transitions int
	var decrements int
	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
				transitions++
				return false, nil // already claimed by a previous delivery
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

	body, sig := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_1", Status: "succeeded"})
	for i := 0; i < 2; i++ {
		if _, err := svc.ProcessWebhook(body, sig); err != nil {
			t.Fatalf("delivery %d: unexpected error: %v", i+1, err)
		}
	}

	if transitions != 2 {
		t.Errorf("expected 2 transition attempts, got %d", transitions)
	}
	if decrements != 0 {
		t.Errorf("expected no stock decrements on replay, got %d", decrements)
	}
}

func TestProcessWebhookFailed(t *testing.T) {
	var failedFrom models.PaymentStatus
	var failedTo models.PaymentStatus
	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
				failedFrom, failedTo = from, to
				return true, nil
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
			return &models.Order{ID: 10, UserID: 1}, nil
		},
	}

	svc := newTestPaymentService(scope, paymentRepo, orderRepo, &stubCartRepo{}, &stubProductRepo{})

	body, sig := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_1", Status: "failed"})
	resp, err := svc.ProcessWebhook(body, sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if failedFrom != models.PaymentStatusPending || failedTo != models.PaymentStatusFailed {
		t.Errorf("expected pending -> failed transition, got %q -> %q", failedFrom, failedTo)
	}
	if resp == nil || resp.Status != "pending" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestProcessWebhookAmountMismatch(t *testing.T) {
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

	body, sig := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_1", Status: "succeeded", Amount: 42})
	_, err := svc.ProcessWebhook(body, sig)
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeWebhookAmountMismatch)
}

func TestProcessWebhookInvalidStatus(t *testing.T) {
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

	body, sig := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_1", Status: "charged_back"})
	_, err := svc.ProcessWebhook(body, sig)
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidWebhookStatus)
}

func TestProcessWebhookPaymentNotFound(t *testing.T) {
	svc := newTestPaymentService(nil, &stubPaymentRepo{}, &stubOrderRepo{}, &stubCartRepo{}, &stubProductRepo{})

	body, sig := signedWebhookBody(t, dto.PaymentWebhookEvent{Reference: "PAY_MISSING", Status: "succeeded"})
	_, err := svc.ProcessWebhook(body, sig)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodePaymentNotFound)
}

func TestSimulateWebhook(t *testing.T) {
	scope := &fakeScope{
		payment: &stubPaymentRepo{
			transitionFn: func(reference string, from, to models.PaymentStatus) (bool, error) {
				return true, nil
			},
		},
		product: &stubProductRepo{},
		cart:    &stubCartRepo{},
		order:   &stubOrderRepo{},
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
			return &models.Cart{ID: 5}, nil
		},
	}

	svc := newTestPaymentService(scope, paymentRepo, orderRepo, cartRepo, &stubProductRepo{})

	resp, err := svc.SimulateWebhook("PAY_1", "succeeded")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Reference != "PAY_1" {
		t.Errorf("unexpected response: %+v", resp)
	}
}
