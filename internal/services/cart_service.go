package services

import (
	"errors"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"
)

type CartService struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository) *CartService {
	return &CartService{cartRepo: cartRepo, productRepo: productRepo}
}

func (s *CartService) GetCart(userID uint) (*models.Cart, error) {
	cart, err := s.cartRepo.GetWithItems(userID)
	if err == nil {
		return cart, nil
	}

	if !errors.Is(err, repositories.ErrRecordNotFound) {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchCart,
			"failed to fetch cart",
			err,
		)
	}

	newCart := &models.Cart{UserID: userID}

	if err := s.cartRepo.Create(newCart); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeCreateCart,
			"failed to create cart",
			err,
		)
	}

	return newCart, nil
}

func (s *CartService) GetCartResponse(userID uint) (*dto.CartResponse, error) {
	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	return mapper.ToCartResponse(cart), nil
}

func (s *CartService) AddToCart(userID uint, req *dto.AddToCartRequest) (*dto.CartResponse, error) {

	if req.Quantity <= 0 {
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidQuantity,
			"quantity must be greater than zero",
			nil,
		)
	}

	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchProduct, "failed to fetch product",
			apperrors.CodeProductNotFound, "product not found",
		)
	}

	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) > 0 {
		existingMerchantID := cart.Items[0].Product.MerchantID

		if existingMerchantID != product.MerchantID {
			return nil, apperrors.New(
				http.StatusConflict,
				apperrors.CodeMultipleMerchants,
				"you can only add products from one merchant to your cart",
				nil,
			)
		}
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
				http.StatusConflict,
				apperrors.CodeInsufficientStock,
				"insufficient stock for the requested quantity",
				nil,
			)
		}

		existing.Quantity = newQty
		existing.Price = product.Price

		if err := s.cartRepo.UpdateItem(existing); err != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeUpdateCartItem,
				"failed to update cart item",
				err,
			)
		}

	} else {
		if product.Stock < req.Quantity {
			return nil, apperrors.New(
				http.StatusConflict,
				apperrors.CodeInsufficientStock,
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
				apperrors.CodeAddCartItem,
				"failed to add item to cart",
				err,
			)
		}
	}

	cart, err = s.recalculateCart(cart.ID, userID)
	if err != nil {
		return nil, err
	}

	return mapper.ToCartResponse(cart), nil
}

func (s *CartService) recalculateCart(cartID uint, userID uint) (*models.Cart, error) {

	cart, err := s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchCart,
			"failed to fetch cart",
			err,
		)
	}

	var total int64
	var count int

	for _, item := range cart.Items {
		total += item.Price * int64(item.Quantity)
		count += item.Quantity
	}

	if err := s.cartRepo.UpdateCartTotal(cartID, total, count); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeUpdateCart,
			"failed to update cart",
			err,
		)
	}

	cart, err = s.cartRepo.GetWithItems(userID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchCart,
			"failed to fetch cart",
			err,
		)
	}
	return cart, nil
}

func (s *CartService) UpdateCartItem(userID, itemID uint, qty int) (*dto.CartResponse, error) {

	if qty <= 0 {
		return nil, apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidQuantity,
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
				return nil, repoErr(
					err,
					apperrors.CodeFetchProduct, "failed to fetch product",
					apperrors.CodeProductNotFound, "product not found",
				)
			}

			if product.Stock < qty {
				return nil, apperrors.New(
					http.StatusConflict,
					apperrors.CodeInsufficientStock,
					"insufficient stock for the requested quantity",
					nil,
				)
			}

			item.Quantity = qty

			if err := s.cartRepo.UpdateItem(item); err != nil {
				return nil, apperrors.New(
					http.StatusInternalServerError,
					apperrors.CodeUpdateCartItem,
					"failed to update cart item",
					err,
				)
			}

			updatedCart, err := s.recalculateCart(cart.ID, userID)
			if err != nil {
				return nil, err
			}

			return mapper.ToCartResponse(updatedCart), nil
		}
	}

	return nil, apperrors.New(
		http.StatusNotFound,
		apperrors.CodeCartItemNotFound,
		"cart item not found",
		nil,
	)
}

func (s *CartService) RemoveFromCart(userID, itemID uint) (*dto.CartResponse, error) {

	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	var existing *models.CartItem

	for i := range cart.Items {
		if cart.Items[i].ID == itemID {
			existing = &cart.Items[i]
			break
		}
	}

	if existing == nil {
		return nil, apperrors.New(
			http.StatusNotFound,
			apperrors.CodeCartItemNotFound,
			"cart item not found",
			nil,
		)
	}

	if err := s.cartRepo.RemoveItem(itemID); err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				apperrors.CodeCartItemNotFound,
				"cart item not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeRemoveCartItem,
			"failed to remove cart item",
			err,
		)
	}

	updatedCart, err := s.recalculateCart(cart.ID, userID)
	if err != nil {
		return nil, err
	}

	return mapper.ToCartResponse(updatedCart), nil
}

func (s *CartService) ClearCart(userID uint) error {
	cart, err := s.GetCart(userID)
	if err != nil {
		return err
	}

	return s.cartRepo.ClearCart(cart.ID)
}
