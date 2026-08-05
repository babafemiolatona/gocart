package services

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/repositories"
)

func newTestCategoryService(repo *stubCategoryRepo) *CategoryService {
	return NewCategoryService(repo)
}

func TestCreateCategorySuccess(t *testing.T) {
	var created *models.Category
	repo := &stubCategoryRepo{
		createFn: func(c *models.Category) error { created = c; return nil },
	}
	svc := newTestCategoryService(repo)

	resp, err := svc.CreateCategory(&dto.CategoryRequest{Name: "Electronics", Description: "Gadgets", Slug: "electronics"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created == nil || created.Name != "Electronics" || created.Description != "Gadgets" || created.Slug != "electronics" {
		t.Errorf("category not created correctly: %+v", created)
	}
	if resp == nil || resp.Name != "Electronics" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCreateCategoryDuplicate(t *testing.T) {
	repo := &stubCategoryRepo{
		createFn: func(c *models.Category) error { return repositories.ErrDuplicate },
	}
	svc := newTestCategoryService(repo)

	_, err := svc.CreateCategory(&dto.CategoryRequest{Name: "Electronics", Slug: "electronics"})
	assertAppError(t, err, http.StatusConflict, apperrors.CodeCategoryExists)
}

func TestGetCategoryByIDNotFound(t *testing.T) {
	svc := newTestCategoryService(&stubCategoryRepo{})

	_, err := svc.GetCategoryByID(99)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeCategoryNotFound)
}

func TestGetCategoryByIDSuccess(t *testing.T) {
	repo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 3, Name: "Books", Slug: "books"}, nil
		},
	}
	svc := newTestCategoryService(repo)

	resp, err := svc.GetCategoryByID(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 3 || resp.Name != "Books" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetAllCategories(t *testing.T) {
	repo := &stubCategoryRepo{
		getAllFn: func() ([]models.Category, error) {
			return []models.Category{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}, nil
		},
	}
	svc := newTestCategoryService(repo)

	resp, err := svc.GetAllCategories()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 categories, got %d", len(resp))
	}
}

func TestUpdateCategorySuccess(t *testing.T) {
	name := "New Name"
	var updatedID uint
	var updatedValues map[string]interface{}

	repo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 3, Name: "Old", Slug: "old"}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error {
			updatedID = id
			updatedValues = values
			return nil
		},
	}
	svc := newTestCategoryService(repo)

	resp, err := svc.UpdateCategory(&dto.UpdateCategoryRequest{Name: &name}, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedID != 3 || updatedValues["name"] != "New Name" {
		t.Errorf("update not applied correctly: id=%d values=%v", updatedID, updatedValues)
	}
	if resp.Name != "New Name" {
		t.Errorf("expected updated name in response, got %q", resp.Name)
	}
}

func TestUpdateCategoryDuplicate(t *testing.T) {
	repo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 3, Name: "Old"}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error { return repositories.ErrDuplicate },
	}
	svc := newTestCategoryService(repo)

	slug := "duplicate"
	_, err := svc.UpdateCategory(&dto.UpdateCategoryRequest{Slug: &slug}, 3)
	assertAppError(t, err, http.StatusConflict, apperrors.CodeCategoryExists)
}

func TestDeleteCategorySuccess(t *testing.T) {
	var deleted uint
	repo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 3, Name: "Books"}, nil
		},
		deleteFn: func(id uint) error { deleted = id; return nil },
	}
	svc := newTestCategoryService(repo)

	if err := svc.DeleteCategory(3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 3 {
		t.Errorf("expected to delete category 3, got %d", deleted)
	}
}

func TestDeleteCategoryNotFound(t *testing.T) {
	svc := newTestCategoryService(&stubCategoryRepo{})

	err := svc.DeleteCategory(99)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeCategoryNotFound)
}

func TestGetAllCategoriesRepoError(t *testing.T) {
	repo := &stubCategoryRepo{
		getAllFn: func() ([]models.Category, error) { return nil, errBoom },
	}
	svc := newTestCategoryService(repo)

	_, err := svc.GetAllCategories()
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchCategories)
}

func TestUpdateCategoryRepoError(t *testing.T) {
	repo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 3, Name: "Old"}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error { return errBoom },
	}
	svc := newTestCategoryService(repo)

	name := "New"
	_, err := svc.UpdateCategory(&dto.UpdateCategoryRequest{Name: &name}, 3)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeUpdateCategory)
}
