package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

type PaymentService struct {
	uow         *repositories.UnitOfWork
	paymentRepo repositories.PaymentRepository
	orderRepo   repositories.OrderRepository
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewPaymentService(
	uow *repositories.UnitOfWork,
	paymentRepo repositories.PaymentRepository,
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	productRepo repositories.ProductRepository,
) *PaymentService {
	return &PaymentService{
		uow:         uow,
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
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

	err = s.uow.WithTransaction(func(uow *repositories.UnitOfWork) error {

		claimed, err := uow.Payment().TransitionStatus(
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

			if err := uow.Product().DecrementStock(item.ProductID, item.Quantity); err != nil {
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

		if err := uow.Cart().ClearCart(cart.ID); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeClearCart,
				"failed to clear cart",
				err,
			)
		}

		if err := uow.Order().UpdateOrderStatus(
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
