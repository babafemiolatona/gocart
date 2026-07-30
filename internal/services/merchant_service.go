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
}

func NewMerchantService(
	merchantRepo repositories.MerchantRepository,
	authRepo repositories.AuthRepository,
) *MerchantService {
	return &MerchantService{
		merchantRepo: merchantRepo,
		userRepo:     authRepo,
	}
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
