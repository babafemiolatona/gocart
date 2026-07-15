package services

import (
	"errors"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"net/http"

	"gorm.io/gorm"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryService struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryService(categoryRepo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) CreateCategory(req *models.CategoryRequest) (*models.Category, error) {
	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				"category_exists",
				"category already exists",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"create_category_failed",
			"failed to create category",
			err,
		)
	}

	return category, nil
}

func (s *CategoryService) GetCategoryByID(id uint) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"category_not_found",
				"category not found",
				err,
			)
		}
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_category_failed",
			"failed to fetch category",
			err,
		)
	}
	return category, nil
}

func (s *CategoryService) GetAllCategories() ([]models.Category, error) {
	categories, err := s.categoryRepo.GetAll()

	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_categories_failed",
			"failed to fetch categories",
			err,
		)
	}

	return categories, nil
}

func (s *CategoryService) UpdateCategory(req *models.CategoryRequest, id uint) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"category_not_found",
				"category not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_category_failed",
			"failed to fetch category",
			err,
		)
	}

	if req.Name != "" {
		category.Name = req.Name
	}

	if req.Description != "" {
		category.Description = req.Description
	}

	if req.Slug != "" {
		category.Slug = req.Slug
	}

	if err := s.categoryRepo.Update(category); err != nil {

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				"category_exists",
				"category already exists",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"update_category_failed",
			"failed to update category",
			err,
		)
	}

	return category, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	_, err := s.categoryRepo.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(
				http.StatusNotFound,
				"category_not_found",
				"category not found",
				err,
			)
		}

		return apperrors.New(
			http.StatusInternalServerError,
			"fetch_category_failed",
			"failed to fetch category",
			err,
		)
	}

	if err := s.categoryRepo.Delete(id); err != nil {
		return apperrors.New(
			http.StatusInternalServerError,
			"delete_category_failed",
			"failed to delete category",
			err,
		)
	}

	return nil
}
