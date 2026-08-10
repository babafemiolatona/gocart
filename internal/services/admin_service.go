package services

import (
	"net/http"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/repositories"
)

type AdminService struct {
	userRepo     repositories.UserRepository
	merchantRepo repositories.MerchantRepository
	productRepo  repositories.ProductRepository
	orderRepo    repositories.OrderRepository
}

func NewAdminService(
	userRepo repositories.UserRepository,
	merchantRepo repositories.MerchantRepository,
	productRepo repositories.ProductRepository,
	orderRepo repositories.OrderRepository,
) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		merchantRepo: merchantRepo,
		productRepo:  productRepo,
		orderRepo:    orderRepo,
	}
}

func (s *AdminService) GetDashboard() (*dto.AdminDashboardResponse, error) {
	totalUsers, err := s.userRepo.CountAll()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchAdminDashboard,
			"failed to fetch admin dashboard",
			err,
		)
	}

	totalMerchants, err := s.merchantRepo.CountAll()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchAdminDashboard,
			"failed to fetch admin dashboard",
			err,
		)
	}

	totalProducts, err := s.productRepo.CountAll()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchAdminDashboard,
			"failed to fetch admin dashboard",
			err,
		)
	}

	totalOrders, err := s.orderRepo.CountAll()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchAdminDashboard,
			"failed to fetch admin dashboard",
			err,
		)
	}

	totalRevenue, err := s.orderRepo.SumRevenueAll()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchAdminDashboard,
			"failed to fetch admin dashboard",
			err,
		)
	}

	ordersByStatus, err := s.orderRepo.CountsByStatus()
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchAdminDashboard,
			"failed to fetch admin dashboard",
			err,
		)
	}

	return &dto.AdminDashboardResponse{
		TotalUsers:     totalUsers,
		TotalMerchants: totalMerchants,
		TotalProducts:  totalProducts,
		TotalOrders:    totalOrders,
		TotalRevenue:   mapper.MinorUnitsToUnit(totalRevenue),
		OrdersByStatus: ordersByStatus,
	}, nil
}
