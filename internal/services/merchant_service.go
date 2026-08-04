package services

import (
	"errors"
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"

	"gorm.io/gorm"
)

type MerchantService struct {
	merchantRepo repositories.MerchantRepository
	userRepo     repositories.AuthRepository
	orderRepo    repositories.OrderRepository
	productRepo  repositories.ProductRepository
}

func NewMerchantService(
	merchantRepo repositories.MerchantRepository,
	authRepo repositories.AuthRepository,
	orderRepo repositories.OrderRepository,
	productRepo repositories.ProductRepository,
) *MerchantService {
	return &MerchantService{
		merchantRepo: merchantRepo,
		userRepo:     authRepo,
		orderRepo:    orderRepo,
		productRepo:  productRepo,
	}
}

func (s *MerchantService) getMerchant(userID uint) (*models.Merchant, error) {
	merchant, err := s.merchantRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"merchant_not_found",
				"merchant profile not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_merchant_failed",
			"failed to fetch merchant profile",
			err,
		)
	}

	return merchant, nil
}

func (s *MerchantService) RegisterMerchant(
	userID uint,
	req *dto.MerchantRegisterRequest,
) (*dto.MerchantResponse, error) {

	var merchant models.Merchant

	err := s.merchantRepo.WithTransaction(func(tx *gorm.DB) error {

		_, err := s.userRepo.GetByIDTx(tx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(
					http.StatusNotFound,
					"user_not_found",
					"user not found",
					err,
				)
			}

			return apperrors.New(
				http.StatusInternalServerError,
				"fetch_user_failed",
				"failed to fetch user",
				err,
			)
		}

		merchant = models.Merchant{
			UserID:       userID,
			BusinessName: req.BusinessName,
			Description:  req.Description,
			Phone:        req.Phone,
			LogoURL:      req.LogoURL,
			IsVerified:   false,
		}

		if err := s.merchantRepo.CreateTx(tx, &merchant); err != nil {

			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return apperrors.New(
					http.StatusConflict,
					"merchant_exists",
					"user is already a merchant",
					err,
				)
			}

			return apperrors.New(
				http.StatusInternalServerError,
				"create_merchant_failed",
				"failed to create merchant",
				err,
			)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mapper.ToMerchantResponse(&merchant), nil
}

func (s *MerchantService) GetProfile(userID uint) (*dto.MerchantResponse, error) {
	merchant, err := s.merchantRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"merchant_not_found",
				"merchant profile not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_merchant_failed",
			"failed to fetch merchant profile",
			err,
		)
	}

	return mapper.ToMerchantResponse(merchant), nil
}

func (s *MerchantService) UpdateProfile(
	userID uint,
	req *dto.UpdateMerchantRequest,
) (*dto.MerchantResponse, error) {

	merchant, err := s.merchantRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"merchant_not_found",
				"merchant profile not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_merchant_failed",
			"failed to fetch merchant profile",
			err,
		)
	}

	if req.BusinessName != nil {
		merchant.BusinessName = *req.BusinessName
	}

	if req.Description != nil {
		merchant.Description = *req.Description
	}

	if req.Phone != nil {
		merchant.Phone = *req.Phone
	}

	if req.LogoURL != nil {
		merchant.LogoURL = *req.LogoURL
	}

	if err := s.merchantRepo.Update(merchant); err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"update_merchant_failed",
			"failed to update merchant profile",
			err,
		)
	}

	return mapper.ToMerchantResponse(merchant), nil
}

func (s *MerchantService) GetOrders(
	userID uint,
) ([]dto.MerchantOrderResponse, error) {

	merchant, err := s.getMerchant(userID)
	if err != nil {
		return nil, err
	}

	orders, err := s.orderRepo.GetOrdersByMerchantID(merchant.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_orders_failed",
			"failed to fetch orders",
			err,
		)
	}

	return mapper.ToMerchantOrderResponses(orders), nil
}

func (s *MerchantService) GetOrder(
	userID uint,
	orderID uint,
) (*dto.MerchantOrderResponse, error) {

	merchant, err := s.getMerchant(userID)
	if err != nil {
		return nil, err
	}

	order, err := s.orderRepo.GetMerchantOrderByID(merchant.ID, orderID)
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

	return mapper.ToMerchantOrderResponse(order), nil
}

func (s *MerchantService) UpdateOrderStatus(
	userID uint,
	orderID uint,
	req *dto.UpdateOrderStatusRequest,
) error {

	merchant, err := s.getMerchant(userID)
	if err != nil {
		return err
	}

	order, err := s.orderRepo.GetMerchantOrderByID(
		merchant.ID,
		orderID,
	)
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

	switch req.Status {

	case models.OrderStatusShipped:

		if order.Status != models.OrderStatusConfirmed {
			return apperrors.New(
				http.StatusBadRequest,
				"invalid_order_status",
				"only confirmed orders can be shipped",
				nil,
			)
		}

	case models.OrderStatusDelivered:

		if order.Status != models.OrderStatusShipped {
			return apperrors.New(
				http.StatusBadRequest,
				"invalid_order_status",
				"only shipped orders can be marked as delivered",
				nil,
			)
		}

	default:

		return apperrors.New(
			http.StatusBadRequest,
			"invalid_order_status",
			"merchant can only update an order to 'shipped' or 'delivered'",
			nil,
		)
	}

	return s.orderRepo.UpdateOrderStatus(orderID, req.Status)
}

func (s *MerchantService) GetDashboard(
	userID uint,
) (*dto.MerchantDashboardResponse, error) {

	merchant, err := s.getMerchant(userID)
	if err != nil {
		return nil, err
	}

	totalProducts, err := s.productRepo.CountByMerchant(merchant.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	lowStockProducts, err := s.productRepo.CountLowStockByMerchant(
		merchant.ID,
		5,
	)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	totalOrders, err := s.orderRepo.CountByMerchant(merchant.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	pendingOrders, err := s.orderRepo.CountByMerchantAndStatus(
		merchant.ID,
		models.OrderStatusConfirmed,
	)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	completedOrders, err := s.orderRepo.CountByMerchantAndStatus(
		merchant.ID,
		models.OrderStatusDelivered,
	)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	totalRevenue, err := s.orderRepo.SumRevenueByMerchant(merchant.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	recentOrders, err := s.orderRepo.GetRecentOrdersByMerchant(
		merchant.ID,
		5,
	)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_dashboard_failed",
			"failed to fetch dashboard",
			err,
		)
	}

	return &dto.MerchantDashboardResponse{
		TotalProducts:    totalProducts,
		TotalOrders:      totalOrders,
		PendingOrders:    pendingOrders,
		CompletedOrders:  completedOrders,
		TotalRevenue:     mapper.MinorUnitsToUnit(totalRevenue),
		LowStockProducts: lowStockProducts,
		RecentOrders:     mapper.ToMerchantRecentOrderResponses(recentOrders),
	}, nil
}
