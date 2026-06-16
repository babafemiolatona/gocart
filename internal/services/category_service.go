package services

import (
	"errors"
	"fmt"
	"gocart/internal/models"
	"gocart/internal/repositories"

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
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) GetCategoryByID(id uint) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) GetAllCategories() ([]models.Category, error) {
	return s.categoryRepo.GetAll()
}

func (s *CategoryService) UpdateCategory(req *models.CategoryRequest, id uint) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(id)

	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	return s.categoryRepo.Delete(id)
}
