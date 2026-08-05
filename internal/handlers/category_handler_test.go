package handlers

import (
	"net/http"
	"testing"

	"gocart/internal/dto"
)

func TestCreateCategorySuccess(t *testing.T) {
	svc := &stubCategoryService{
		createFn: func(req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
			return &dto.CategoryResponse{ID: 1, Name: req.Name, Slug: req.Slug}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/admin/categories", NewCategoryHandler(svc).CreateCategory)

	w := doRequest(t, r, http.MethodPost, "/admin/categories", `{"name":"Electronics","slug":"electronics"}`)
	assertStatus(t, w, http.StatusCreated)

	var resp dto.CategoryResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 1 || resp.Name != "Electronics" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCreateCategoryValidationError(t *testing.T) {
	svc := &stubCategoryService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/admin/categories", NewCategoryHandler(svc).CreateCategory)

	w := doRequest(t, r, http.MethodPost, "/admin/categories", `{}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetCategoriesSuccess(t *testing.T) {
	svc := &stubCategoryService{
		getAllFn: func() ([]dto.CategoryResponse, error) {
			return []dto.CategoryResponse{{ID: 1, Name: "Electronics"}, {ID: 2, Name: "Books"}}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/categories", NewCategoryHandler(svc).GetCategories)

	w := doRequest(t, r, http.MethodGet, "/categories", "")
	assertStatus(t, w, http.StatusOK)

	var resp []dto.CategoryResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 categories, got %+v", resp)
	}
}

func TestGetCategoryByIDSuccess(t *testing.T) {
	svc := &stubCategoryService{
		getByIDFn: func(id uint) (*dto.CategoryResponse, error) {
			return &dto.CategoryResponse{ID: id, Name: "Books"}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/categories/:id", NewCategoryHandler(svc).GetCategoryByID)

	w := doRequest(t, r, http.MethodGet, "/categories/3", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetCategoryByIDInvalidID(t *testing.T) {
	svc := &stubCategoryService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/categories/:id", NewCategoryHandler(svc).GetCategoryByID)

	w := doRequest(t, r, http.MethodGet, "/categories/abc", "")
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_category_id" {
		t.Errorf("expected invalid_category_id, got %q", code)
	}
}

func TestGetCategoryByIDNotFound(t *testing.T) {
	svc := &stubCategoryService{
		getByIDFn: func(id uint) (*dto.CategoryResponse, error) {
			return nil, appErr(http.StatusNotFound, "category_not_found", "category not found")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/categories/:id", NewCategoryHandler(svc).GetCategoryByID)

	w := doRequest(t, r, http.MethodGet, "/categories/99", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestUpdateCategorySuccess(t *testing.T) {
	svc := &stubCategoryService{
		updateFn: func(req *dto.UpdateCategoryRequest, id uint) (*dto.CategoryResponse, error) {
			if id != 3 {
				t.Errorf("expected id 3, got %d", id)
			}
			return &dto.CategoryResponse{ID: 3}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPut, "/admin/categories/:id", NewCategoryHandler(svc).UpdateCategory)

	w := doRequest(t, r, http.MethodPut, "/admin/categories/3", `{"name":"Books"}`)
	assertStatus(t, w, http.StatusOK)
}

func TestUpdateCategoryInvalidID(t *testing.T) {
	svc := &stubCategoryService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPut, "/admin/categories/:id", NewCategoryHandler(svc).UpdateCategory)

	w := doRequest(t, r, http.MethodPut, "/admin/categories/abc", `{}`)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestDeleteCategorySuccess(t *testing.T) {
	svc := &stubCategoryService{
		deleteFn: func(id uint) error { return nil },
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodDelete, "/admin/categories/:id", NewCategoryHandler(svc).DeleteCategory)

	w := doRequest(t, r, http.MethodDelete, "/admin/categories/3", "")
	assertStatus(t, w, http.StatusNoContent)
}

func TestDeleteCategoryInvalidID(t *testing.T) {
	svc := &stubCategoryService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodDelete, "/admin/categories/:id", NewCategoryHandler(svc).DeleteCategory)

	w := doRequest(t, r, http.MethodDelete, "/admin/categories/xyz", "")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCreateCategoryServiceError(t *testing.T) {
	svc := &stubCategoryService{
		createFn: func(req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
			return nil, appErr(http.StatusConflict, "category_exists", "exists")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPost, "/admin/categories", NewCategoryHandler(svc).CreateCategory)

	w := doRequest(t, r, http.MethodPost, "/admin/categories", `{"name":"Electronics","slug":"electronics"}`)
	assertStatus(t, w, http.StatusConflict)
}

func TestGetCategoriesServiceError(t *testing.T) {
	svc := &stubCategoryService{
		getAllFn: func() ([]dto.CategoryResponse, error) {
			return nil, appErr(http.StatusInternalServerError, "fetch_categories_failed", "failed")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/categories", NewCategoryHandler(svc).GetCategories)

	w := doRequest(t, r, http.MethodGet, "/categories", "")
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestUpdateCategoryServiceError(t *testing.T) {
	svc := &stubCategoryService{
		updateFn: func(req *dto.UpdateCategoryRequest, id uint) (*dto.CategoryResponse, error) {
			return nil, appErr(http.StatusConflict, "category_exists", "exists")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodPut, "/admin/categories/:id", NewCategoryHandler(svc).UpdateCategory)

	w := doRequest(t, r, http.MethodPut, "/admin/categories/3", `{"name":"Books"}`)
	assertStatus(t, w, http.StatusConflict)
}

func TestDeleteCategoryServiceError(t *testing.T) {
	svc := &stubCategoryService{
		deleteFn: func(id uint) error {
			return appErr(http.StatusNotFound, "category_not_found", "not found")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodDelete, "/admin/categories/:id", NewCategoryHandler(svc).DeleteCategory)

	w := doRequest(t, r, http.MethodDelete, "/admin/categories/3", "")
	assertStatus(t, w, http.StatusNotFound)
}
