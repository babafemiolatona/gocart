package services

import (
	"errors"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"

	"github.com/google/uuid"
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
			Reference: uuid.NewString(),
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
		Payment: mapper.ToPaymentCheckoutResponse(payment),
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

func (s *OrderService) GetOrder(orderID uint) (*dto.OrderDetailsResponse, error) {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusNotFound,
			"order_not_found",
			"order not found",
			err,
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

func (s *OrderService) CancelOrder(orderID uint) error {
	err := s.orderRepo.UpdateOrderStatus(orderID, models.OrderStatusCancelled)

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
			"cancel_order_failed",
			"failed to cancel order",
			err,
		)
	}

	return nil
}
