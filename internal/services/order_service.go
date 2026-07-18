package services

import (
	"errors"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"

	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo   repositories.OrderRepository
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewOrderService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	productRepo repositories.ProductRepository,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
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

func (s *OrderService) ProcessCheckout(userID uint, shippingAddress string) (*models.Order, error) {
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
		Status:          models.OrderStatusPending,
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

	err = s.orderRepo.WithTransaction(func(tx *gorm.DB) error {

		if err := s.orderRepo.CreateOrderTx(tx, order); err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"create_order_failed",
				"failed to create order",
				err,
			)
		}

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

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]models.OrderResponse, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_orders_failed",
			"failed to fetch orders",
			err,
		)
	}

	response := make([]models.OrderResponse, len(orders))
	for i, order := range orders {
		response[i] = models.OrderResponse{
			ID:              order.ID,
			Status:          string(order.Status),
			Total:           order.Total,
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt,
		}
	}

	return response, nil

}

func (s *OrderService) GetOrder(orderID uint) (*models.OrderDetailsResponse, error) {
	order, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusNotFound,
			"order_not_found",
			"order not found",
			err,
		)
	}

	response := &models.OrderDetailsResponse{
		ID:              order.ID,
		Status:          string(order.Status),
		Total:           order.Total,
		ShippingAddress: order.ShippingAddress,
		Items:           order.Items,
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
