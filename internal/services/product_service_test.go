package services

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/query"
	"gocart/internal/repositories"
)

func newTestProductService(
	productRepo *stubProductRepo,
	categoryRepo *stubCategoryRepo,
	imageRepo *stubProductImageRepo,
) *ProductService {
	if imageRepo == nil {
		imageRepo = &stubProductImageRepo{}
	}
	return NewProductService(productRepo, categoryRepo, imageRepo, &stubStorage{}, 10*1024*1024)
}

func newImageHeader(t *testing.T, name string) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("image", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("fake-image-bytes")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	reader := multipart.NewReader(&buf, w.Boundary())
	form, err := reader.ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	defer form.RemoveAll()

	files := form.File["image"]
	if len(files) == 0 {
		t.Fatal("no file parts parsed")
	}
	h := files[0]
	h.Header = textproto.MIMEHeader{"Content-Type": {"image/png"}}
	return h
}

func TestCreateProductSuccess(t *testing.T) {
	var created *models.Product
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error {
			created = p
			p.ID = 1
			return nil
		},
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, Name: "Laptop", Price: 149999, Stock: 5}, nil
		},
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 2, Name: "Electronics"}, nil
		},
	}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	resp, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Description: "Fast", Price: 1499.99, Stock: 5,
		CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created == nil || created.MerchantID != 1 || created.Price != 149999 {
		t.Errorf("product not created correctly: %+v", created)
	}
	if resp == nil || resp.Name != "Laptop" || resp.Price != 1499.99 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCreateProductCategoryNotFound(t *testing.T) {
	productRepo := &stubProductRepo{}
	categoryRepo := &stubCategoryRepo{}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 99, Sku: "LAP-1", Slug: "laptop",
	}, nil)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeCategoryNotFound)
}

func TestCreateProductDuplicate(t *testing.T) {
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error { return repositories.ErrDuplicate },
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 2}, nil
		},
	}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, nil)
	assertAppError(t, err, http.StatusConflict, apperrors.CodeProductExists)
}

func TestGetProductNotFound(t *testing.T) {
	svc := newTestProductService(&stubProductRepo{}, &stubCategoryRepo{}, nil)

	_, err := svc.GetProduct(99)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeProductNotFound)
}

func TestGetProductsAppliesDefaults(t *testing.T) {
	productRepo := &stubProductRepo{
		getAllFn: func(q *dto.PaginationQuery, filters *query.ProductFilters) ([]models.Product, int64, error) {
			if q.Page != 1 || q.PageSize != 10 {
				t.Errorf("expected default pagination query, got %+v", q)
			}
			return nil, 0, nil
		},
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	resp, err := svc.GetProducts(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Page != 1 || resp.PageSize != 10 {
		t.Errorf("expected default pagination, got %+v", resp)
	}
}

func TestGetProductsPaginates(t *testing.T) {
	productRepo := &stubProductRepo{
		getAllFn: func(q *dto.PaginationQuery, filters *query.ProductFilters) ([]models.Product, int64, error) {
			return []models.Product{{ID: 1}, {ID: 2}, {ID: 3}}, 3, nil
		},
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	resp, err := svc.GetProducts(&dto.PaginationQuery{Page: 1, PageSize: 2}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Page != 1 || resp.PageSize != 2 || resp.Total != 3 || resp.TotalPage != 2 {
		t.Errorf("unexpected pagination: %+v", resp)
	}
}

func TestUpdateProductNotOwner(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 2}, nil
		},
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	price := 20.0
	_, err := svc.UpdateProduct(1, 7, &dto.UpdateProductRequest{Price: &price}, nil)
	assertAppError(t, err, http.StatusForbidden, apperrors.CodeForbidden)
}

func TestUpdateProductSuccess(t *testing.T) {
	var updatedValues map[string]interface{}
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error {
			updatedValues = values
			return nil
		},
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) {
			return &models.Category{ID: 2}, nil
		},
	}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	stock := 10
	resp, err := svc.UpdateProduct(1, 1, &dto.UpdateProductRequest{Stock: &stock}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedValues["stock"] != 10 {
		t.Errorf("expected stock update, got %v", updatedValues)
	}
	if resp == nil {
		t.Error("expected a response")
	}
}

func TestUpdateProductCategoryNotFound(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
	}
	categoryRepo := &stubCategoryRepo{}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	catID := uint(99)
	_, err := svc.UpdateProduct(1, 1, &dto.UpdateProductRequest{CategoryID: &catID}, nil)
	assertAppError(t, err, http.StatusNotFound, apperrors.CodeCategoryNotFound)
}

func TestDeleteProductNotOwner(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 2}, nil
		},
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	err := svc.DeleteProduct(1, 1)
	assertAppError(t, err, http.StatusForbidden, apperrors.CodeForbidden)
}

func TestDeleteProductSuccess(t *testing.T) {
	var deleted uint
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		deleteFn: func(id uint) error { deleted = id; return nil },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	if err := svc.DeleteProduct(1, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected to delete product 1, got %d", deleted)
	}
}

func TestGetMerchantProductNotOwned(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 2}, nil
		},
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	_, err := svc.GetMerchantProduct(1, 1)
	assertAppError(t, err, http.StatusForbidden, apperrors.CodeAccessDenied)
}

func TestCreateProductImageTooLarge(t *testing.T) {
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error { p.ID = 1; return nil },
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{ID: 2}, nil },
	}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	image := newImageHeader(t, "photo.png")
	image.Size = 10*1024*1024 + 1

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, []*multipart.FileHeader{image})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeFileTooLarge)
}

func TestCreateProductInvalidImageType(t *testing.T) {
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error { p.ID = 1; return nil },
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{ID: 2}, nil },
	}
	svc := newTestProductService(productRepo, categoryRepo, nil)

	image := newImageHeader(t, "photo.png")
	image.Header = textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}}

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, []*multipart.FileHeader{image})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeInvalidImageType)
}

func TestCreateProductWithImagesSuccess(t *testing.T) {
	var uploaded uint
	var saved []models.ProductImage
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error { p.ID = 1; return nil },
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, Name: "Laptop", Price: 1000}, nil
		},
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{ID: 2}, nil },
	}
	imageRepo := &stubProductImageRepo{
		createManyFn: func(images []models.ProductImage) error {
			saved = images
			return nil
		},
	}
	storage := &stubStorage{
		uploadFn: func(file multipart.File, header *multipart.FileHeader, productID uint) (string, error) {
			uploaded = productID
			return "/images/laptop.png", nil
		},
	}
	svc := NewProductService(productRepo, categoryRepo, imageRepo, storage, 10*1024*1024)

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, []*multipart.FileHeader{newImageHeader(t, "photo.png")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uploaded != 1 {
		t.Errorf("expected upload for product 1, got %d", uploaded)
	}
	if len(saved) != 1 || saved[0].ImageURL != "/images/laptop.png" {
		t.Errorf("unexpected saved images: %+v", saved)
	}
}

func TestCreateProductImageUploadFails(t *testing.T) {
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error { p.ID = 1; return nil },
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{ID: 2}, nil },
	}
	storage := &stubStorage{
		uploadFn: func(file multipart.File, header *multipart.FileHeader, productID uint) (string, error) {
			return "", errBoom
		},
	}
	svc := NewProductService(productRepo, categoryRepo, &stubProductImageRepo{}, storage, 10*1024*1024)

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, []*multipart.FileHeader{newImageHeader(t, "photo.png")})
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeUploadImages)
}

func TestCreateProductImagesSaveRollsBack(t *testing.T) {
	var deleted []string
	productRepo := &stubProductRepo{
		createFn: func(p *models.Product) error { p.ID = 1; return nil },
	}
	categoryRepo := &stubCategoryRepo{
		getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{ID: 2}, nil },
	}
	imageRepo := &stubProductImageRepo{
		createManyFn: func(images []models.ProductImage) error { return errBoom },
	}
	storage := &stubStorage{
		uploadFn: func(file multipart.File, header *multipart.FileHeader, productID uint) (string, error) {
			return "/images/a.png", nil
		},
		deleteFn: func(objectName string) error {
			deleted = append(deleted, objectName)
			return nil
		},
	}
	svc := NewProductService(productRepo, categoryRepo, imageRepo, storage, 10*1024*1024)

	_, err := svc.CreateProduct(1, &dto.CreateProductRequest{
		Name: "Laptop", Price: 10, CategoryID: 2, Sku: "LAP-1", Slug: "laptop",
	}, []*multipart.FileHeader{newImageHeader(t, "photo.png")})
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeUploadImages)

	if len(deleted) != 1 || deleted[0] != "/images/a.png" {
		t.Errorf("expected uploaded object to be deleted, got %v", deleted)
	}
}

func TestGetProductRepoError(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) { return nil, errBoom },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	_, err := svc.GetProduct(1)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchProduct)
}

func TestGetMerchantProductRepoError(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) { return nil, errBoom },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	_, err := svc.GetMerchantProduct(1, 1)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchProduct)
}

func TestDeleteProductRepoError(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) { return nil, errBoom },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	err := svc.DeleteProduct(1, 1)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeFetchProduct)
}

func TestDeleteProductDeleteFails(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		deleteFn: func(id uint) error { return errBoom },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	err := svc.DeleteProduct(1, 1)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeDeleteProduct)
}

func TestDeleteProductStorageErrorStillSucceeds(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{
				ID: 1, MerchantID: 1,
				Images: []models.ProductImage{{ImageURL: "/img/a.png"}},
			}, nil
		},
		deleteFn: func(id uint) error { return nil },
	}
	storage := &stubStorage{
		deleteFn: func(objectName string) error { return errBoom },
	}
	svc := NewProductService(productRepo, &stubCategoryRepo{}, &stubProductImageRepo{}, storage, 10*1024*1024)

	if err := svc.DeleteProduct(1, 1); err != nil {
		t.Fatalf("expected delete to succeed despite storage error, got %v", err)
	}
}

func TestUpdateProductDuplicate(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error { return repositories.ErrDuplicate },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	sku := "DUP"
	_, err := svc.UpdateProduct(1, 1, &dto.UpdateProductRequest{Sku: &sku}, nil)
	assertAppError(t, err, http.StatusConflict, apperrors.CodeProductExists)
}

func TestUpdateProductRepoError(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error { return errBoom },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	sku := "NEW"
	_, err := svc.UpdateProduct(1, 1, &dto.UpdateProductRequest{Sku: &sku}, nil)
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeUpdateProduct)
}

func TestUpdateProductImageTooLarge(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error { return nil },
	}
	svc := newTestProductService(productRepo, &stubCategoryRepo{}, nil)

	image := newImageHeader(t, "photo.png")
	image.Size = 10*1024*1024 + 1

	_, err := svc.UpdateProduct(1, 1, &dto.UpdateProductRequest{}, []*multipart.FileHeader{image})
	assertAppError(t, err, http.StatusBadRequest, apperrors.CodeFileTooLarge)
}

func TestUpdateProductImageUploadFails(t *testing.T) {
	productRepo := &stubProductRepo{
		getByIDFn: func(id uint) (*models.Product, error) {
			return &models.Product{ID: 1, MerchantID: 1}, nil
		},
		updateFn: func(id uint, values map[string]interface{}) error { return nil },
	}
	storage := &stubStorage{
		uploadFn: func(file multipart.File, header *multipart.FileHeader, productID uint) (string, error) {
			return "", errBoom
		},
	}
	svc := NewProductService(productRepo, &stubCategoryRepo{}, &stubProductImageRepo{}, storage, 10*1024*1024)

	_, err := svc.UpdateProduct(1, 1, &dto.UpdateProductRequest{}, []*multipart.FileHeader{newImageHeader(t, "photo.png")})
	assertAppError(t, err, http.StatusInternalServerError, apperrors.CodeUploadImages)
}
