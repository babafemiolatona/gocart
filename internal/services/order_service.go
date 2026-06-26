package services

import (
	"errors"
	"fmt"
	"gocart/internal/models"
	"gocart/internal/repositories"
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
			return fmt.Errorf("product not found: %w", err)
		}

		if product.Stock < item.Quantity {
			return fmt.Errorf("Insufficient stock for product %d", product.ID)
		}
	}
	return nil
}

func (s *OrderService) ProcessCheckout(userID uint, shippingAddress string) (*models.Order, error) {
	cart, err := s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
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
		orderItem := models.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.Product.Name,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}

		order.Items = append(order.Items, orderItem)
	}

	if err := s.orderRepo.CreateOrder(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range cart.Items {

		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch product %d: %w", item.ProductID, err)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %d", product.ID)
		}

		product.Stock -= item.Quantity

		if err := s.productRepo.Update(product); err != nil {
			return nil, fmt.Errorf("failed to update stock for product %d: %w", product.ID, err)
		}
	}

	if err := s.cartRepo.ClearCart(cart.ID); err != nil {
		return nil, fmt.Errorf("order created but failed to clear cart: %w", err)
	}

	return order, nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]models.OrderResponse, error) {
	orders, err := s.orderRepo.GetOrdersByUserID(userID)
	if err != nil {
		return nil, err
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
		return nil, err
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
	return s.orderRepo.UpdateOrderStatus(orderID, models.OrderStatusCancelled)
}
