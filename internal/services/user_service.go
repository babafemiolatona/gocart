package services

import (
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetMe(userID uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchUser, "failed to fetch user",
			apperrors.CodeUserNotFound, "user not found",
		)
	}

	return mapper.ToUserResponse(user), nil
}
