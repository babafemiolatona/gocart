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

const (
	lowStockThreshold    = 5
	recentOrdersLimit    = 5
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
		return nil, repoErr(
			err,
			apperrors.CodeFetchMerchant, "failed to fetch merchant profile",
			apperrors.CodeMerchantNotFound, "merchant profile not found",
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
					apperrors.CodeUserNotFound,
					"user not found",
					err,
				)
			}

			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeFetchUser,
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
					apperrors.CodeMerchantExists,
					"user is already a merchant",
					err,
				)
			}

			return apperrors.New(
				http.StatusInternalServerError,
				apperrors.CodeCreateMerchant,
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
		return nil, repoErr(
			err,
			apperrors.CodeFetchMerchant, "failed to fetch merchant profile",
			apperrors.CodeMerchantNotFound, "merchant profile not found",
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
		return nil, repoErr(
			err,
			apperrors.CodeFetchMerchant, "failed to fetch merchant profile",
			apperrors.CodeMerchantNotFound, "merchant profile not found",
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
			apperrors.CodeUpdateMerchant,
			"failed to update merchant profile",
			err,
		)
	}

	return mapper.ToMerchantResponse(merchant), nil
}

func (s *MerchantService) GetOrders(
	userID uint,
	p *dto.PaginationQuery,
) (*dto.PaginatedResponse, error) {

	merchant, err := s.getMerchant(userID)
	if err != nil {
		return nil, err
	}

	orders, total, err := s.orderRepo.GetOrdersByMerchantID(merchant.ID, p)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchOrders,
			"failed to fetch orders",
			err,
		)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize > 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		Data:      mapper.ToMerchantOrderResponses(orders),
		Total:     total,
		Page:      p.Page,
		PageSize:  p.PageSize,
		TotalPage: totalPages,
	}, nil
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
		return nil, repoErr(
			err,
			apperrors.CodeFetchOrder, "failed to fetch order",
			apperrors.CodeOrderNotFound, "order not found",
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
		return repoErr(
			err,
			apperrors.CodeFetchOrder, "failed to fetch order",
			apperrors.CodeOrderNotFound, "order not found",
		)
	}

	switch req.Status {

	case models.OrderStatusShipped:

		if order.Status != models.OrderStatusConfirmed {
			return apperrors.New(
				http.StatusBadRequest,
				apperrors.CodeInvalidOrderStatus,
				"only confirmed orders can be shipped",
				nil,
			)
		}

	case models.OrderStatusDelivered:

		if order.Status != models.OrderStatusShipped {
			return apperrors.New(
				http.StatusBadRequest,
				apperrors.CodeInvalidOrderStatus,
				"only shipped orders can be marked as delivered",
				nil,
			)
		}

	default:

		return apperrors.New(
			http.StatusBadRequest,
			apperrors.CodeInvalidOrderStatus,
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
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	lowStockProducts, err := s.productRepo.CountLowStockByMerchant(
		merchant.ID,
		lowStockThreshold,
	)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	totalOrders, err := s.orderRepo.CountByMerchant(merchant.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
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
			apperrors.CodeFetchDashboard,
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
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	totalRevenue, err := s.orderRepo.SumRevenueByMerchant(merchant.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	recentOrders, err := s.orderRepo.GetRecentOrdersByMerchant(
		merchant.ID,
		recentOrdersLimit,
	)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
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
