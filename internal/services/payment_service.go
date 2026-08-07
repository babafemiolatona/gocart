package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"gocart/internal/security"
)

type PaymentService struct {
	uow           repositories.TransactionManager
	paymentRepo   repositories.PaymentRepository
	orderRepo     repositories.OrderRepository
	cartRepo      repositories.CartRepository
	productRepo   repositories.ProductRepository
	webhookSecret []byte
}

func NewPaymentService(
	uow repositories.TransactionManager,
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	productRepo repositories.ProductRepository,
	webhookSecret string,
) *PaymentService {
	return &PaymentService{
		uow:           uow,
		paymentRepo:   paymentRepo,
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		productRepo:   productRepo,
		webhookSecret: []byte(webhookSecret),
	}
}

func (s *PaymentService) ProcessPayment(userID uint, reference string) (*dto.PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByReference(reference)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchPayment, "failed to fetch payment",
			apperrors.CodePaymentNotFound, "payment not found",
		)
	}

	order, err := s.orderRepo.GetOrderByID(payment.OrderID)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchOrder, "failed to fetch order",
			apperrors.CodeOrderNotFound, "order not found",
		)
	}

	if order.UserID != userID {
		return nil, apperrors.New(
			http.StatusNotFound,
			apperrors.CodePaymentNotFound,
			"payment not found",
			nil,
		)
	}

	// Idempotency
	if payment.Status == models.PaymentStatusSucceeded {
		return mapper.ToPaymentResponse(payment), nil
	}

	cart, err := s.cartRepo.GetWithItems(order.UserID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchCart,
			"failed to fetch cart",
			err,
		)
	}

	err = s.uow.WithTransaction(func(scope repositories.TransactionScope) error {
		return s.finalizeSucceeded(scope, reference, cart, order)
	})

	if err != nil {
		return nil, err
	}

	payment, err = s.paymentRepo.GetByReference(reference)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchPayment, "failed to fetch payment",
			apperrors.CodePaymentNotFound, "payment not found",
		)
	}

	return mapper.ToPaymentResponse(payment), nil
}

// finalizeSucceeded runs the payment-confirmation side effects inside a single
// transaction: claim the pending payment, deduct stock, clear the cart, and
// confirm the order. The TransitionStatus claim makes replays idempotent, so
// this is safe to call from both the user-facing ProcessPayment and the
// provider webhook.
func (s *PaymentService) finalizeSucceeded(
	scope repositories.TransactionScope,
	reference string,
	cart *models.Cart,
	order *models.Order,
) error {
	claimed, err := scope.Payment().TransitionStatus(
		reference,
		models.PaymentStatusPending,
		models.PaymentStatusSucceeded,
	)
	if err != nil {
		return apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeUpdatePayment,
			"failed to update payment",
			err,
		)
	}

	if !claimed {
		return nil
	}

	for _, item := range cart.Items {

		if err := scope.Product().DecrementStock(item.ProductID, item.Quantity); err != nil {
			if errors.Is(err, repositories.ErrRecordNotFound) {
				return apperrors.New(
					http.StatusNotFound,
					apperrors.CodeProductNotFound,
					"product not found",
					err,
				)
			}

			if errors.Is(err, repositories.ErrInsufficientStock) {
				return apperrors.New(
					http.StatusConflict,
					apperrors.CodeInsufficientStock,
					"insufficient stock",
					err,
				)
			}

			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeUpdateProduct,
				"failed to update product",
				err,
			)
		}
	}

	if err := scope.Cart().ClearCart(cart.ID); err != nil {
		return apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeClearCart,
			"failed to clear cart",
			err,
		)
	}

	if err := scope.Order().UpdateOrderStatus(
		order.ID,
		models.OrderStatusConfirmed,
	); err != nil {
		return apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeUpdateOrder,
			"failed to update order",
			err,
		)
	}

	return nil
}

// ProcessWebhook handles a signed payment-provider callback. The webhook is the
// source of truth for payment outcomes: a succeeded event finalizes the order
// (stock, cart, order status), a failed event marks the payment failed. It is
// idempotent — replays of an already-processed event are safe no-ops.
func (s *PaymentService) ProcessWebhook(body []byte, signature string) (*dto.PaymentResponse, error) {
	if !security.VerifyWebhookSignature(s.webhookSecret, body, signature) {
		return nil, apperrors.New(
			http.StatusUnauthorized,
			apperrors.CodeInvalidWebhookSignature,
			"invalid webhook signature",
			nil,
		)
	}

	var event dto.PaymentWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidWebhookEvent,
			"invalid webhook payload",
			err,
		)
	}

	if event.Reference == "" {
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidWebhookEvent,
			"payment reference is required",
			nil,
		)
	}

	payment, err := s.paymentRepo.GetByReference(event.Reference)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchPayment, "failed to fetch payment",
			apperrors.CodePaymentNotFound, "payment not found",
		)
	}

	order, err := s.orderRepo.GetOrderByID(payment.OrderID)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchOrder, "failed to fetch order",
			apperrors.CodeOrderNotFound, "order not found",
		)
	}

	if event.Amount > 0 && event.Amount != payment.Amount {
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeWebhookAmountMismatch,
			"webhook amount does not match payment",
			nil,
		)
	}

	switch event.Status {
	case string(models.PaymentStatusSucceeded):
		cart, err := s.cartRepo.GetWithItems(order.UserID)
		if err != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeFetchCart,
				"failed to fetch cart",
				err,
			)
		}

		err = s.uow.WithTransaction(func(scope repositories.TransactionScope) error {
			return s.finalizeSucceeded(scope, event.Reference, cart, order)
		})
		if err != nil {
			return nil, err
		}

	case string(models.PaymentStatusFailed):
		err = s.uow.WithTransaction(func(scope repositories.TransactionScope) error {
			if _, err := scope.Payment().TransitionStatus(
				event.Reference,
				models.PaymentStatusPending,
				models.PaymentStatusFailed,
			); err != nil {
				return apperrors.New(
					http.StatusInternalServerError,
					apperrors.CodeUpdatePayment,
					"failed to update payment",
					err,
				)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

	default:
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidWebhookStatus,
			"unsupported webhook status",
			nil,
		)
	}

	payment, err = s.paymentRepo.GetByReference(event.Reference)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchPayment, "failed to fetch payment",
			apperrors.CodePaymentNotFound, "payment not found",
		)
	}

	return mapper.ToPaymentResponse(payment), nil
}

// SimulateWebhook builds a signed webhook event and runs it through
// ProcessWebhook. Dev-only helper that mimics a payment provider callback so
// the full flow can be exercised without an external service.
func (s *PaymentService) SimulateWebhook(reference string, status string) (*dto.PaymentResponse, error) {
	body, err := json.Marshal(dto.PaymentWebhookEvent{
		Reference: reference,
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeInternalServer,
			"failed to build webhook event",
			err,
		)
	}

	return s.ProcessWebhook(body, security.SignWebhookBody(s.webhookSecret, body))
}

func (s *PaymentService) GetPayment(userID uint, reference string) (*dto.PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByReference(reference)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchPayment, "failed to fetch payment",
			apperrors.CodePaymentNotFound, "payment not found",
		)
	}

	order, err := s.orderRepo.GetOrderByID(payment.OrderID)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchPayment, "failed to fetch payment",
			apperrors.CodePaymentNotFound, "payment not found",
		)
	}

	if order.UserID != userID {
		return nil, apperrors.New(
			http.StatusNotFound,
			apperrors.CodePaymentNotFound,
			"payment not found",
			nil,
		)
	}

	return mapper.ToPaymentResponse(payment), nil
}

func generateReference() (string, error) {
	bytes := make([]byte, 8)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return "PAY_" + hex.EncodeToString(bytes), nil
}
