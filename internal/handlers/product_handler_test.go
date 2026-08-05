package handlers

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"gocart/internal/dto"
	"gocart/internal/query"
)

func TestGetProductsSuccess(t *testing.T) {
	var gotQuery *dto.PaginationQuery
	svc := &stubProductService{
		getProductsFn: func(q *dto.PaginationQuery, f *query.ProductFilters) (*dto.PaginatedResponse, error) {
			gotQuery = q
			return &dto.PaginatedResponse{Total: 2, Page: q.Page, PageSize: q.PageSize, TotalPage: 1}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/products", NewProductHandler(svc).GetProducts)

	w := doRequest(t, r, http.MethodGet, "/products?page=2&page_size=5", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.PaginatedResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.Total != 2 || resp.Page != 2 || resp.PageSize != 5 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if gotQuery == nil || gotQuery.Page != 2 || gotQuery.PageSize != 5 {
		t.Errorf("unexpected query parsed: %+v", gotQuery)
	}
}

func TestGetProductSuccess(t *testing.T) {
	svc := &stubProductService{
		getProductFn: func(id uint) (*dto.ProductResponse, error) {
			return &dto.ProductResponse{ID: id, Name: "Laptop"}, nil
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/products/:id", NewProductHandler(svc).GetProduct)

	w := doRequest(t, r, http.MethodGet, "/products/3", "")
	assertStatus(t, w, http.StatusOK)

	var resp dto.ProductResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 3 || resp.Name != "Laptop" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetProductInvalidID(t *testing.T) {
	svc := &stubProductService{}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/products/:id", NewProductHandler(svc).GetProduct)

	w := doRequest(t, r, http.MethodGet, "/products/abc", "")
	assertStatus(t, w, http.StatusBadRequest)

	code, _ := decodeError(t, w)
	if code != "invalid_product_id" {
		t.Errorf("expected invalid_product_id, got %q", code)
	}
}

func TestGetProductNotFound(t *testing.T) {
	svc := &stubProductService{
		getProductFn: func(id uint) (*dto.ProductResponse, error) {
			return nil, appErr(http.StatusNotFound, "product_not_found", "product not found")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/products/:id", NewProductHandler(svc).GetProduct)

	w := doRequest(t, r, http.MethodGet, "/products/99", "")
	assertStatus(t, w, http.StatusNotFound)
}

func TestCreateProductSuccess(t *testing.T) {
	var gotMerchant uint
	svc := &stubProductService{
		createFn: func(merchantID uint, req *dto.CreateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error) {
			gotMerchant = merchantID
			return &dto.ProductResponse{ID: 1, Name: req.Name}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPost, "/merchants/products", NewProductHandler(svc).CreateProduct)

	body, contentType := buildMultipartForm(t, map[string]string{
		"name":        "Laptop",
		"description": "Fast",
		"price":       "1499.99",
		"stock":       "5",
		"category_id": "2",
		"sku":         "LAP-1",
		"slug":        "laptop",
	})

	req := newRequestWithBody(t, http.MethodPost, "/merchants/products", body, contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusCreated)

	var resp dto.ProductResponse
	decodeJSON(t, w.Body.Bytes(), &resp)
	if resp.ID != 1 || resp.Name != "Laptop" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if gotMerchant != 2 {
		t.Errorf("expected merchant 2, got %d", gotMerchant)
	}
}

func TestCreateProductValidationError(t *testing.T) {
	svc := &stubProductService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPost, "/merchants/products", NewProductHandler(svc).CreateProduct)

	w := doRequest(t, r, http.MethodPost, "/merchants/products", "")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestUpdateProductSuccess(t *testing.T) {
	var gotMerchant, gotID uint
	svc := &stubProductService{
		updateFn: func(merchantID, id uint, req *dto.UpdateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error) {
			gotMerchant, gotID = merchantID, id
			return &dto.ProductResponse{ID: id}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPut, "/merchants/products/:id", NewProductHandler(svc).UpdateProduct)

	w := doRequest(t, r, http.MethodPut, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusOK)

	if gotMerchant != 2 || gotID != 5 {
		t.Errorf("unexpected args: merchant=%d id=%d", gotMerchant, gotID)
	}
}

func TestUpdateProductInvalidID(t *testing.T) {
	svc := &stubProductService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPut, "/merchants/products/:id", NewProductHandler(svc).UpdateProduct)

	w := doRequest(t, r, http.MethodPut, "/merchants/products/abc", "")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestDeleteProductSuccess(t *testing.T) {
	var gotMerchant, gotID uint
	svc := &stubProductService{
		deleteFn: func(merchantID, id uint) error {
			gotMerchant, gotID = merchantID, id
			return nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodDelete, "/merchants/products/:id", NewProductHandler(svc).DeleteProduct)

	w := doRequest(t, r, http.MethodDelete, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusNoContent)
	if gotMerchant != 2 || gotID != 5 {
		t.Errorf("unexpected args: merchant=%d id=%d", gotMerchant, gotID)
	}
}

func TestDeleteProductInvalidID(t *testing.T) {
	svc := &stubProductService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodDelete, "/merchants/products/:id", NewProductHandler(svc).DeleteProduct)

	w := doRequest(t, r, http.MethodDelete, "/merchants/products/xyz", "")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetMerchantProducts(t *testing.T) {
	svc := &stubProductService{
		getProductsFn: func(q *dto.PaginationQuery, f *query.ProductFilters) (*dto.PaginatedResponse, error) {
			if f.MerchantID != 2 {
				t.Errorf("expected merchant filter 2, got %d", f.MerchantID)
			}
			return &dto.PaginatedResponse{Total: 1, Page: 1, PageSize: 10, TotalPage: 1}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/products", NewProductHandler(svc).GetMerchantProducts)

	w := doRequest(t, r, http.MethodGet, "/merchants/products", "")
	assertStatus(t, w, http.StatusOK)
}

func TestGetMerchantProduct(t *testing.T) {
	var gotMerchant, gotID uint
	svc := &stubProductService{
		getMerchantFn: func(merchantID, id uint) (*dto.ProductResponse, error) {
			gotMerchant, gotID = merchantID, id
			return &dto.ProductResponse{ID: id}, nil
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/products/:id", NewProductHandler(svc).GetMerchantProduct)

	w := doRequest(t, r, http.MethodGet, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusOK)
	if gotMerchant != 2 || gotID != 5 {
		t.Errorf("unexpected args: merchant=%d id=%d", gotMerchant, gotID)
	}
}

func TestCreateProductMerchantUnauthorized(t *testing.T) {
	svc := &stubProductService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPost, "/merchants/products", NewProductHandler(svc).CreateProduct)

	body, contentType := buildMultipartForm(t, map[string]string{"name": "Laptop", "price": "1", "stock": "1", "category_id": "1", "sku": "S", "slug": "s"})
	req := newRequestWithBody(t, http.MethodPost, "/merchants/products", body, contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusUnauthorized)
}

func TestCreateProductServiceError(t *testing.T) {
	svc := &stubProductService{
		createFn: func(merchantID uint, req *dto.CreateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error) {
			return nil, appErr(http.StatusConflict, "product_exists", "exists")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPost, "/merchants/products", NewProductHandler(svc).CreateProduct)

	body, contentType := buildMultipartForm(t, map[string]string{"name": "Laptop", "price": "1", "stock": "1", "category_id": "1", "sku": "S", "slug": "s"})
	req := newRequestWithBody(t, http.MethodPost, "/merchants/products", body, contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertStatus(t, w, http.StatusConflict)
}

func TestGetProductsServiceError(t *testing.T) {
	svc := &stubProductService{
		getProductsFn: func(q *dto.PaginationQuery, f *query.ProductFilters) (*dto.PaginatedResponse, error) {
			return nil, appErr(http.StatusInternalServerError, "fetch_products_failed", "failed")
		},
	}
	r := newTestRouter(0, nil)
	registerHandler(r, http.MethodGet, "/products", NewProductHandler(svc).GetProducts)

	w := doRequest(t, r, http.MethodGet, "/products", "")
	assertStatus(t, w, http.StatusInternalServerError)
}

func TestUpdateProductMerchantUnauthorized(t *testing.T) {
	svc := &stubProductService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodPut, "/merchants/products/:id", NewProductHandler(svc).UpdateProduct)

	w := doRequest(t, r, http.MethodPut, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestUpdateProductServiceError(t *testing.T) {
	svc := &stubProductService{
		updateFn: func(merchantID, id uint, req *dto.UpdateProductRequest, images []*multipart.FileHeader) (*dto.ProductResponse, error) {
			return nil, appErr(http.StatusForbidden, "forbidden", "not owner")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodPut, "/merchants/products/:id", NewProductHandler(svc).UpdateProduct)

	w := doRequest(t, r, http.MethodPut, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusForbidden)
}

func TestDeleteProductMerchantUnauthorized(t *testing.T) {
	svc := &stubProductService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodDelete, "/merchants/products/:id", NewProductHandler(svc).DeleteProduct)

	w := doRequest(t, r, http.MethodDelete, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestDeleteProductServiceError(t *testing.T) {
	svc := &stubProductService{
		deleteFn: func(merchantID, id uint) error {
			return appErr(http.StatusForbidden, "forbidden", "not owner")
		},
	}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodDelete, "/merchants/products/:id", NewProductHandler(svc).DeleteProduct)

	w := doRequest(t, r, http.MethodDelete, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusForbidden)
}

func TestGetMerchantProductsUnauthorized(t *testing.T) {
	svc := &stubProductService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/merchants/products", NewProductHandler(svc).GetMerchantProducts)

	w := doRequest(t, r, http.MethodGet, "/merchants/products", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetMerchantProductUnauthorized(t *testing.T) {
	svc := &stubProductService{}
	r := newTestRouter(7, nil)
	registerHandler(r, http.MethodGet, "/merchants/products/:id", NewProductHandler(svc).GetMerchantProduct)

	w := doRequest(t, r, http.MethodGet, "/merchants/products/5", "")
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetMerchantProductInvalidID(t *testing.T) {
	svc := &stubProductService{}
	mid := uint(2)
	r := newTestRouter(7, &mid)
	registerHandler(r, http.MethodGet, "/merchants/products/:id", NewProductHandler(svc).GetMerchantProduct)

	w := doRequest(t, r, http.MethodGet, "/merchants/products/abc", "")
	assertStatus(t, w, http.StatusBadRequest)
}
