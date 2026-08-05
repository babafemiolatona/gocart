package services

import (
	"errors"
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

const (
	lowStockThreshold = 5
	recentOrdersLimit = 5
)

type MerchantService struct {
	uow          *repositories.UnitOfWork
	merchantRepo repositories.MerchantRepository
	orderRepo    repositories.OrderRepository
	productRepo  repositories.ProductRepository
}

func NewMerchantService(
	uow *repositories.UnitOfWork,
	merchantRepo repositories.MerchantRepository,
	orderRepo repositories.OrderRepository,
	productRepo repositories.ProductRepository,
) *MerchantService {
	return &MerchantService{
		uow:          uow,
		merchantRepo: merchantRepo,
		orderRepo:    orderRepo,
		productRepo:  productRepo,
	}
}

func (s *MerchantService) RegisterMerchant(
	userID uint,
	req *dto.MerchantRegisterRequest,
) (*dto.MerchantResponse, error) {

	var merchant models.Merchant

	err := s.uow.WithTransaction(func(uow *repositories.UnitOfWork) error {

		_, err := uow.Auth().GetByID(userID)
		if err != nil {
			if errors.Is(err, repositories.ErrRecordNotFound) {
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

		if err := uow.Merchant().Create(&merchant); err != nil {

			if errors.Is(err, repositories.ErrDuplicate) {
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

func (s *MerchantService) GetProfile(merchantID uint) (*dto.MerchantResponse, error) {
	merchant, err := s.merchantRepo.GetByID(merchantID)
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
	merchantID uint,
	req *dto.UpdateMerchantRequest,
) (*dto.MerchantResponse, error) {

	merchant, err := s.merchantRepo.GetByID(merchantID)
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
	merchantID uint,
	p *dto.PaginationQuery,
) (*dto.PaginatedResponse, error) {

	orders, total, err := s.orderRepo.GetOrdersByMerchantID(merchantID, p)
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
	merchantID uint,
	orderID uint,
) (*dto.MerchantOrderResponse, error) {

	order, err := s.orderRepo.GetMerchantOrderByID(merchantID, orderID)
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
	merchantID uint,
	orderID uint,
	req *dto.UpdateOrderStatusRequest,
) error {

	order, err := s.orderRepo.GetMerchantOrderByID(
		merchantID,
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
	merchantID uint,
) (*dto.MerchantDashboardResponse, error) {

	totalProducts, err := s.productRepo.CountByMerchant(merchantID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	lowStockProducts, err := s.productRepo.CountLowStockByMerchant(
		merchantID,
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

	totalOrders, err := s.orderRepo.CountByMerchant(merchantID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	awaitingShipment, err := s.orderRepo.CountByMerchantAndStatus(
		merchantID,
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
		merchantID,
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

	totalRevenue, err := s.orderRepo.SumRevenueByMerchant(merchantID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchDashboard,
			"failed to fetch dashboard",
			err,
		)
	}

	recentOrders, err := s.orderRepo.GetRecentOrdersByMerchant(
		merchantID,
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
		AwaitingShipment: awaitingShipment,
		CompletedOrders:  completedOrders,
		TotalRevenue:     mapper.MinorUnitsToUnit(totalRevenue),
		LowStockProducts: lowStockProducts,
		RecentOrders:     mapper.ToMerchantRecentOrderResponses(recentOrders),
	}, nil
}
