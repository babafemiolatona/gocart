package services

import (
	"errors"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"

	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo   repositories.OrderRepository
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
	paymentRepo repositories.PaymentRepository
}

var errCheckoutConflict = errors.New("checkout idempotency conflict")

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	productRepo repositories.ProductRepository,
	paymentRepo repositories.PaymentRepository,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *OrderService) ValidateCart(cart *models.Cart) error {
	for _, item := range cart.Items {
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(
					http.StatusNotFound,
					apperrors.CodeProductNotFound,
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
				http.StatusBadRequest,
				"insufficient_stock",
				"insufficient stock",
				nil,
			)
		}
	}
	return nil
}

func (s *OrderService) ProcessCheckout(
	userID uint,
	shippingAddress string,
	idempotencyKey string,
) (*dto.CheckoutResponse, error) {

	if idempotencyKey != "" {
		existing, err := s.orderRepo.GetByUserIDAndIdempotencyKey(userID, idempotencyKey)
		if err == nil {
			payment, err := s.paymentRepo.GetByOrderID(existing.ID)
			if err == nil {
				return &dto.CheckoutResponse{
					Order:   mapper.ToOrderCheckoutResponse(existing),
					Payment: mapper.ToPaymentResponse(payment),
				}, nil
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeFetchOrder,
				"failed to fetch order",
				err,
			)
		}
	}

	cart, err := s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchCart,
			"failed to fetch cart",
			err,
		)
	}

	if len(cart.Items) == 0 {
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeCartEmpty,
			"cart is empty",
			nil,
		)
	}

	if err := s.ValidateCart(cart); err != nil {
		return nil, err
	}

	order := &models.Order{
		UserID:          userID,
		Status:          models.OrderStatusPendingPayment,
		Total:           cart.Total,
		ShippingAddress: shippingAddress,
		IdempotencyKey:  idempotencyKey,
	}

	for _, item := range cart.Items {
		order.Items = append(order.Items, models.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.Product.Name,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	reference, err := generateReference()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeGenerateReference,
			"failed to generate payment reference",
			err,
		)
	}

	var payment *models.Payment

	err = s.orderRepo.WithTransaction(func(tx *gorm.DB) error {

		if err := s.orderRepo.CreateOrderTx(tx, order); err != nil {
			if idempotencyKey != "" && errors.Is(err, gorm.ErrDuplicatedKey) {
				return errCheckoutConflict
			}

			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeCreateOrder,
				"failed to create order",
				err,
			)
		}

		payment = &models.Payment{
			OrderID:        order.ID,
			Reference:      reference,
			Amount:         order.Total,
			IdempotencyKey: idempotencyKey,
			Status:         models.PaymentStatusPending,
			Provider:       "mock",
		}

		if err := s.paymentRepo.CreateTx(tx, payment); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeCreatePayment,
				"failed to create payment",
				err,
			)
		}

		return nil
	})

	if errors.Is(err, errCheckoutConflict) {
		existing, fetchErr := s.orderRepo.GetByUserIDAndIdempotencyKey(userID, idempotencyKey)
		if fetchErr != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeFetchOrder,
				"failed to fetch order",
				fetchErr,
			)
		}

		existingPayment, fetchErr := s.paymentRepo.GetByOrderID(existing.ID)
		if fetchErr != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				"fetch_payment_failed",
				"failed to fetch payment",
				fetchErr,
			)
		}

		return &dto.CheckoutResponse{
			Order:   mapper.ToOrderCheckoutResponse(existing),
			Payment: mapper.ToPaymentResponse(existingPayment),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return &dto.CheckoutResponse{
		Order:   mapper.ToOrderCheckoutResponse(order),
		Payment: mapper.ToPaymentResponse(payment),
	}, nil
}

func (s *OrderService) GetUserOrders(userID uint, p *dto.PaginationQuery) (*dto.PaginatedResponse, error) {
	orders, total, err := s.orderRepo.GetOrdersByUserID(userID, p)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchOrders,
			"failed to fetch orders",
			err,
		)
	}

	response := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		response[i] = dto.OrderResponse{
			ID:              order.ID,
			Status:          string(order.Status),
			Total:           mapper.MinorUnitsToUnit(order.Total),
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt,
		}
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize > 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		Data:      response,
		Total:     total,
		Page:      p.Page,
		PageSize:  p.PageSize,
		TotalPage: totalPages,
	}, nil

}

func (s *OrderService) GetOrder(userID, orderID uint) (*dto.OrderDetailsResponse, error) {
	order, err := s.orderRepo.GetOrderByID(orderID)
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
			apperrors.CodeOrderNotFound,
			"order not found",
			nil,
		)
	}

	response := &dto.OrderDetailsResponse{
		ID:              order.ID,
		Status:          string(order.Status),
		Total:           mapper.MinorUnitsToUnit(order.Total),
		ShippingAddress: order.ShippingAddress,
		Items:           mapper.ToOrderItemResponses(order.Items),
		CreatedAt:       order.CreatedAt,
	}

	return response, nil
}

func (s *OrderService) CancelOrder(userID, orderID uint) error {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return repoErr(
			err,
			apperrors.CodeFetchOrder, "failed to fetch order",
			apperrors.CodeOrderNotFound, "order not found",
		)
	}

	if order.UserID != userID {
		return apperrors.New(
			http.StatusNotFound,
			apperrors.CodeOrderNotFound,
			"order not found",
			nil,
		)
	}

	if order.Status == models.OrderStatusCancelled {
		return apperrors.New(
			http.StatusConflict,
			apperrors.CodeOrderAlreadyClosed,
			"order has already been cancelled",
			nil,
		)
	}

	if order.Status == models.OrderStatusPendingPayment {
		if err := s.orderRepo.UpdateOrderStatus(
			order.ID,
			models.OrderStatusCancelled,
		); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeCancelOrder,
				"failed to cancel order",
				err,
			)
		}

		return nil
	}

	if order.Status == models.OrderStatusConfirmed {
		err := s.orderRepo.WithTransaction(func(tx *gorm.DB) error {

			for _, item := range order.Items {

				if err := s.productRepo.IncrementStockTx(tx, item.ProductID, item.Quantity); err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return apperrors.New(
							http.StatusNotFound,
							apperrors.CodeProductNotFound,
							"product not found",
							err,
						)
					}

					return apperrors.New(
						http.StatusInternalServerError,
						"update_product_failed",
						"failed to update product",
						err,
					)
				}
			}

			if err := s.orderRepo.UpdateOrderStatusTx(
				tx,
				order.ID,
				models.OrderStatusCancelled,
			); err != nil {
				return apperrors.New(
					http.StatusInternalServerError,
					apperrors.CodeCancelOrder,
					"failed to cancel order",
					err,
				)
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	}

	return apperrors.New(
		http.StatusConflict,
		apperrors.CodeInvalidOrderStatus,
		"order cannot be cancelled",
		nil,
	)
}
