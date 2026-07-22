package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"

	"gorm.io/gorm"
)

type PaymentService struct {
	paymentRepo repositories.PaymentRepository
	orderRepo   repositories.OrderRepository
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	productRepo repositories.ProductRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *PaymentService) InitiatePayment(orderID uint) (*models.Payment, error) {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"order_not_found",
				"order not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_order_failed",
			"failed to fetch order",
			err,
		)
	}

	reference, err := generateReference()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"generate_reference_failed",
			"failed to generate payment reference",
			err,
		)
	}

	payment := &models.Payment{
		OrderID:   order.ID,
		Reference: reference,
		Amount:    order.Total,
		Status:    models.PaymentStatusPending,
		Provider:  "mock",
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"create_payment_failed",
			"failed to create payment",
			err,
		)
	}

	return payment, nil
}

func (s *PaymentService) ProcessPayment(reference string) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByReference(reference)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"payment_not_found",
				"payment not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_payment_failed",
			"failed to fetch payment",
			err,
		)
	}

	// Idempotency
	if payment.Status == models.PaymentStatusSucceeded {
		return payment, nil
	}

	order, err := s.orderRepo.GetOrderByID(payment.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"order_not_found",
				"order not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_order_failed",
			"failed to fetch order",
			err,
		)
	}

	cart, err := s.cartRepo.GetWithItems(order.UserID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_cart_failed",
			"failed to fetch cart",
			err,
		)
	}

	err = s.orderRepo.WithTransaction(func(tx *gorm.DB) error {

		for _, item := range cart.Items {

			product, err := s.productRepo.GetByIDTx(tx, item.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.New(
						http.StatusNotFound,
						"product_not_found",
						"product not found",
						err,
					)
				}

				return apperrors.New(
					http.StatusInternalServerError,
					"fetch_product_failed",
					"failed to fetch product",
					err,
				)
			}

			if product.Stock < item.Quantity {
				return apperrors.New(
					http.StatusConflict,
					"insufficient_stock",
					"insufficient stock",
					nil,
				)
			}

			product.Stock -= item.Quantity

			if err := s.productRepo.UpdateTx(tx, product); err != nil {
				return apperrors.New(
					http.StatusInternalServerError,
					"update_product_failed",
					"failed to update product",
					err,
				)
			}
		}

		if err := s.cartRepo.ClearCartTx(tx, cart.ID); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"clear_cart_failed",
				"failed to clear cart",
				err,
			)
		}

		if err := s.paymentRepo.UpdateStatusTx(
			tx,
			reference,
			models.PaymentStatusSucceeded,
		); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"update_payment_failed",
				"failed to update payment",
				err,
			)
		}

		if err := s.orderRepo.UpdateOrderStatusTx(
			tx,
			order.ID,
			models.OrderStatusConfirmed,
		); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"update_order_failed",
				"failed to update order",
				err,
			)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.paymentRepo.GetByReference(reference)
}

func (s *PaymentService) GetPayment(reference string) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByReference(reference)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"payment_not_found",
				"payment not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_payment_failed",
			"failed to fetch payment",
			err,
		)
	}

	return payment, nil
}

func generateReference() (string, error) {
	bytes := make([]byte, 8)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return "PAY_" + hex.EncodeToString(bytes), nil
}
