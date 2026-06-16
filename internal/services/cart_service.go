package services

import (
	"errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

type CartService struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewCartService(cartRepo repositories.CartRepository, prouctRepo repositories.ProductRepository) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: prouctRepo}
}

func (s *CartService) GetCart(userID uint) (*models.Cart, error) {
	cart, err := s.cartRepo.GetWithItems(userID)
	if err == nil {
		return cart, nil
	}

	// create if not exists
	newCart := &models.Cart{UserID: userID}

	if err := s.cartRepo.Create(newCart); err != nil {
		return nil, err
	}

	return s.cartRepo.GetWithItems(userID)
}

func (s *CartService) AddToCart(userID uint, req *models.AddToCartRequest) (*models.Cart, error) {

	if req.Quantity <= 0 {
		return nil, errors.New("invalid quantity")
	}

	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	var existing *models.CartItem

	for i := range cart.Items {
		if cart.Items[i].ProductID == req.ProductID {
			existing = &cart.Items[i]
			break
		}
	}

	if existing != nil {
		newQty := existing.Quantity + req.Quantity

		if product.Stock < newQty {
			return nil, errors.New("insufficient stock")
		}

		existing.Quantity = newQty
		existing.Price = product.Price

		if err := s.cartRepo.UpdateItem(existing); err != nil {
			return nil, err
		}

	} else {
		if product.Stock < req.Quantity {
			return nil, errors.New("insufficient stock")
		}

		newItem := &models.CartItem{
			CartID:    cart.ID,
			ProductID: product.ID,
			Quantity:  req.Quantity,
			Price:     product.Price,
		}

		if err := s.cartRepo.AddItem(newItem); err != nil {
			return nil, err
		}
	}

	return s.recalculateCart(cart.ID, userID)
}

func (s *CartService) recalculateCart(cartID uint, userID uint) (*models.Cart, error) {

	cart, err := s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, err
	}

	var total float64
	var count int

	for _, item := range cart.Items {
		total += float64(item.Quantity) * item.Price
		count += item.Quantity
	}

	if err := s.cartRepo.UpdateCartTotal(cartID, total, count); err != nil {
		return nil, err
	}

	return s.cartRepo.GetWithItems(userID)
}

func (s *CartService) UpdateCartItem(userID, itemID uint, qty int) (*models.Cart, error) {

	if qty <= 0 {
		return nil, errors.New("invalid quantity")
	}

	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	for i := range cart.Items {
		item := &cart.Items[i]

		if item.ID == itemID {

			product, err := s.productRepo.GetByID(item.ProductID)
			if err != nil {
				return nil, errors.New("product not found")
			}

			if product.Stock < qty {
				return nil, errors.New("insufficient stock")
			}

			item.Quantity = qty

			if err := s.cartRepo.UpdateItem(item); err != nil {
				return nil, err
			}

			return s.recalculateCart(cart.ID, userID)
		}
	}

	return nil, errors.New("cart item not found")
}

func (s *CartService) RemoveFromCart(userID, itemID uint) (*models.Cart, error) {

	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.RemoveItem(cart.ID); err != nil {
		return nil, err
	}

	return s.recalculateCart(cart.ID, userID)
}

func (s *CartService) ClearCart(userID uint) error {
	cart, err := s.GetCart(userID)
	if err != nil {
		return err
	}

	return s.cartRepo.ClearCart(cart.ID)
}
