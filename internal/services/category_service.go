package services

import (
	"errors"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"

	"gorm.io/gorm"
)

type CategoryService struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryService(categoryRepo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) CreateCategory(req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				apperrors.CodeCategoryExists,
				"category already exists",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeCreateCategory,
			"failed to create category",
			err,
		)
	}

	return mapper.ToCategoryResponse(category), nil
}

func (s *CategoryService) GetCategoryByID(id uint) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchCategory, "failed to fetch category",
			apperrors.CodeCategoryNotFound, "category not found",
		)
	}
	return mapper.ToCategoryResponse(category), nil
}

func (s *CategoryService) GetAllCategories() ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetAll()

	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeFetchCategories,
			"failed to fetch categories",
			err,
		)
	}

	return mapper.ToCategoryResponses(categories), nil
}

func (s *CategoryService) UpdateCategory(req *dto.UpdateCategoryRequest, id uint) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, repoErr(
			err,
			apperrors.CodeFetchCategory, "failed to fetch category",
			apperrors.CodeCategoryNotFound, "category not found",
		)
	}

	updates := map[string]interface{}{}

	if req.Name != nil {
		category.Name = *req.Name
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		category.Description = *req.Description
		updates["description"] = *req.Description
	}

	if req.Slug != nil {
		category.Slug = *req.Slug
		updates["slug"] = *req.Slug
	}

	if err := s.categoryRepo.Update(id, updates); err != nil {

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				apperrors.CodeCategoryExists,
				"category already exists",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeUpdateCategory,
			"failed to update category",
			err,
		)
	}

	return mapper.ToCategoryResponse(category), nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	_, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return repoErr(
			err,
			apperrors.CodeFetchCategory, "failed to fetch category",
			apperrors.CodeCategoryNotFound, "category not found",
		)
	}

	if err := s.categoryRepo.Delete(id); err != nil {
		return apperrors.New(
			http.StatusInternalServerError,
			apperrors.CodeDeleteCategory,
			"failed to delete category",
			err,
		)
	}

	return nil
}
