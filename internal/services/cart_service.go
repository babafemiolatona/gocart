package services

import (
	"errors"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"

	"gorm.io/gorm"
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

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_cart_failed",
			"failed to fetch cart",
			err,
		)
	}

	newCart := &models.Cart{UserID: userID}

	if err := s.cartRepo.Create(newCart); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"create_cart_failed",
			"failed to create cart",
			err,
		)
	}

	cart, err = s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_cart_failed",
			"failed to fetch cart",
			err,
		)
	}

	return cart, nil
}

func (s *CartService) AddToCart(userID uint, req *dto.AddToCartRequest) (*models.Cart, error) {

	if req.Quantity <= 0 {
		return nil, apperrors.New(
			http.StatusBadRequest,
			"invalid_quantity",
			"quantity must be greater than zero",
			nil,
		)
	}

	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"product_not_found",
				"product not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_product_failed",
			"failed to fetch product",
			err,
		)
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
			return nil, apperrors.New(
				http.StatusBadRequest,
				"insufficient_stock",
				"insufficient stock for the requested quantity",
				nil,
			)
		}

		existing.Quantity = newQty
		existing.Price = product.Price

		if err := s.cartRepo.UpdateItem(existing); err != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				"update_cart_item_failed",
				"failed to update cart item",
				err,
			)
		}

	} else {
		if product.Stock < req.Quantity {
			return nil, apperrors.New(
				http.StatusConflict,
				"insufficient_stock",
				"insufficient stock for the requested quantity",
				nil,
			)
		}

		newItem := &models.CartItem{
			CartID:    cart.ID,
			ProductID: product.ID,
			Quantity:  req.Quantity,
			Price:     product.Price,
		}

		if err := s.cartRepo.AddItem(newItem); err != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				"add_cart_item_failed",
				"failed to add item to cart",
				err,
			)
		}
	}

	return s.recalculateCart(cart.ID, userID)
}

func (s *CartService) recalculateCart(cartID uint, userID uint) (*models.Cart, error) {

	cart, err := s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_cart_failed",
			"failed to fetch cart",
			err,
		)
	}

	var total float64
	var count int

	for _, item := range cart.Items {
		total += float64(item.Quantity) * item.Price
		count += item.Quantity
	}

	if err := s.cartRepo.UpdateCartTotal(cartID, total, count); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"update_cart_failed",
			"failed to update cart",
			err,
		)
	}

	cart, err = s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_cart_failed",
			"failed to fetch cart",
			err,
		)
	}
	return cart, nil
}

func (s *CartService) UpdateCartItem(userID, itemID uint, qty int) (*models.Cart, error) {

	if qty <= 0 {
		return nil, apperrors.New(
			http.StatusBadRequest,
			"invalid_quantity",
			"quantity must be a positive integer",
			nil,
		)
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
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, apperrors.New(
						http.StatusNotFound,
						"product_not_found",
						"product not found",
						err,
					)
				}

				return nil, apperrors.New(
					http.StatusInternalServerError,
					"fetch_product_failed",
					"failed to fetch product",
					err,
				)
			}

			if product.Stock < qty {
				return nil, apperrors.New(
					http.StatusConflict,
					"insufficient_stock",
					"insufficient stock for the requested quantity",
					nil,
				)
			}

			item.Quantity = qty

			if err := s.cartRepo.UpdateItem(item); err != nil {
				return nil, apperrors.New(
					http.StatusInternalServerError,
					"update_cart_item_failed",
					"failed to update cart item",
					err,
				)
			}

			return s.recalculateCart(cart.ID, userID)
		}
	}

	return nil, apperrors.New(
		http.StatusNotFound,
		"cart_item_not_found",
		"cart item not found",
		nil,
	)
}

func (s *CartService) RemoveFromCart(userID, itemID uint) (*models.Cart, error) {

	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	found := false

	for _, item := range cart.Items {
		if item.ID == itemID {
			found = true
			break
		}
	}

	if !found {
		return nil, apperrors.New(
			http.StatusNotFound,
			"cart_item_not_found",
			"cart item not found",
			nil,
		)
	}

	if err := s.cartRepo.RemoveItem(itemID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"cart_item_not_found",
				"cart item not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"remove_cart_item_failed",
			"failed to remove cart item",
			err,
		)
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
