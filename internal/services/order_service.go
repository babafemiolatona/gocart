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
) (*dto.CheckoutResponse, error) {

	cart, err := s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_cart_failed",
			"failed to fetch cart",
			err,
		)
	}

	if len(cart.Items) == 0 {
		return nil, apperrors.New(
			http.StatusBadRequest,
			"cart_empty",
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
			"generate_reference_failed",
			"failed to generate payment reference",
			err,
		)
	}

	var payment *models.Payment

	err = s.orderRepo.WithTransaction(func(tx *gorm.DB) error {

		if err := s.orderRepo.CreateOrderTx(tx, order); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"create_order_failed",
				"failed to create order",
				err,
			)
		}

		payment = &models.Payment{
			OrderID:   order.ID,
			Reference: reference,
			Amount:    order.Total,
			Status:    models.PaymentStatusPending,
			Provider:  "mock",
		}

		if err := s.paymentRepo.CreateTx(tx, payment); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"create_payment_failed",
				"failed to create payment",
				err,
			)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.CheckoutResponse{
		Order:   mapper.ToOrderCheckoutResponse(order),
		Payment: mapper.ToPaymentResponse(payment),
	}, nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_orders_failed",
			"failed to fetch orders",
			err,
		)
	}

	response := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		response[i] = dto.OrderResponse{
			ID:              order.ID,
			Status:          string(order.Status),
			Total:           order.Total,
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt,
		}
	}

	return response, nil

}

func (s *OrderService) GetOrder(userID, orderID uint) (*dto.OrderDetailsResponse, error) {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusNotFound,
			"order_not_found",
			"order not found",
			err,
		)
	}

	if order.UserID != userID {
		return nil, apperrors.New(
			http.StatusNotFound,
			"order_not_found",
			"order not found",
			nil,
		)
	}

	response := &dto.OrderDetailsResponse{
		ID:              order.ID,
		Status:          string(order.Status),
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
		Items:           mapper.ToOrderItemResponses(order.Items),
		CreatedAt:       order.CreatedAt,
	}

	return response, nil
}

func (s *OrderService) CancelOrder(userID, orderID uint) error {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(
				http.StatusNotFound,
				"order_not_found",
				"order not found",
				err,
			)
		}

		return apperrors.New(
			http.StatusInternalServerError,
			"fetch_order_failed",
			"failed to fetch order",
			err,
		)
	}

	if order.UserID != userID {
		return apperrors.New(
			http.StatusNotFound,
			"order_not_found",
			"order not found",
			nil,
		)
	}

	if order.Status == models.OrderStatusCancelled {
		return apperrors.New(
			http.StatusConflict,
			"order_already_cancelled",
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
				"cancel_order_failed",
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
							"product_not_found",
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
					"cancel_order_failed",
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
		"invalid_order_status",
		"order cannot be cancelled",
		nil,
	)
}
