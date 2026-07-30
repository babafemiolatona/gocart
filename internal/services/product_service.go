package services

import (
	"errors"
	"fmt"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/mapper"
	"gocart/internal/query"
	"net/http"

	"gocart/internal/models"
	"gocart/internal/repositories"
	"gocart/internal/storage"
	"mime/multipart"

	"gorm.io/gorm"
)

type ProductService struct {
	productRepo      repositories.ProductRepository
	categoryRepo     repositories.CategoryRepository
	productImageRepo repositories.ProductImageRepository
	storage          storage.Storage
}

func NewProductService(
	productRepo repositories.ProductRepository,
	categoryRepo repositories.CategoryRepository,
	productImageRepo repositories.ProductImageRepository,
	storage storage.Storage,
) *ProductService {
	return &ProductService{
		productRepo:      productRepo,
		categoryRepo:     categoryRepo,
		productImageRepo: productImageRepo,
		storage:          storage,
	}
}

func (s *ProductService) CreateProduct(
	merchantID uint,
	req *dto.CreateProductRequest,
	images []*multipart.FileHeader,
) (*dto.ProductResponse, error) {

	_, err := s.categoryRepo.GetByID(req.CategoryID)
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

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		MerchantID:  merchantID,
		Sku:         req.Sku,
		Slug:        req.Slug,
	}

	if err := s.productRepo.Create(product); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				"product_exists",
				"product already exists",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"create_product_failed",
			"failed to create product",
			err,
		)
	}

	if len(images) > 0 {
		if err := s.uploadImages(product.ID, images); err != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				"upload_product_images_failed",
				"failed to upload product images",
				err,
			)
		}
	}

	product, err = s.productRepo.GetByID(product.ID)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_product_failed",
			"failed to fetch product",
			err,
		)
	}

	return mapper.ToProductResponse(product), nil
}

func (s *ProductService) uploadImages(
	productID uint,
	images []*multipart.FileHeader,
) error {

	productImages := make([]models.ProductImage, 0, len(images))

	for _, image := range images {

		file, err := image.Open()
		if err != nil {
			return fmt.Errorf("failed to open image: %w", err)
		}

		objectName, err := s.storage.UploadProductImage(
			file,
			image,
			productID,
		)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to upload image: %w", err)
		}

		file.Close()

		productImages = append(productImages, models.ProductImage{
			ProductID: productID,
			ImageURL:  objectName,
		})
	}

	if len(productImages) > 0 {
		if err := s.productImageRepo.CreateMany(productImages); err != nil {
			return fmt.Errorf("failed to save product images: %w", err)
		}
	}

	return nil
}

func (s *ProductService) GetProduct(id uint) (*dto.ProductResponse, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"product_not_found",
				"product not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_product_failed",
			"failed to fetch product",
			err,
		)
	}

	return mapper.ToProductResponse(product), nil
}

func (s *ProductService) GetProducts(query *dto.PaginationQuery, filters *query.ProductFilters) (*dto.PaginatedResponse, error) {

	if query == nil {
		query = &dto.PaginationQuery{
			Page:     1,
			PageSize: 10,
			Sort:     "created_at",
			Order:    "desc",
		}
	}

	if query.Page < 1 {
		query.Page = 1
	}

	if query.PageSize < 1 {
		query.PageSize = 10
	}

	products, total, err := s.productRepo.GetAll(query, filters)
	if err != nil {
		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_products_failed",
			"failed to fetch products",
			err,
		)
	}

	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		Data:      mapper.ToProductResponses(products),
		Total:     total,
		Page:      query.Page,
		PageSize:  query.PageSize,
		TotalPage: totalPages,
	}, nil
}

func (s *ProductService) UpdateProduct(
	id uint,
	merchantID uint,
	req *dto.UpdateProductRequest,
	images []*multipart.FileHeader,
) (*dto.ProductResponse, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"product_not_found",
				"product not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_product_failed",
			"failed to fetch product",
			err,
		)
	}

	if product.MerchantID != merchantID {
		return nil, apperrors.New(
			http.StatusForbidden,
			"forbidden",
			"you do not own this product",
			nil,
		)
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Sku != nil {
		product.Sku = *req.Sku
	}
	if req.Slug != nil {
		product.Slug = *req.Slug
	}

	if err := s.productRepo.Update(product); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, apperrors.New(
				http.StatusConflict,
				"product_exists",
				"product already exists",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"update_product_failed",
			"failed to update product",
			err,
		)
	}

	if len(images) > 0 {
		if err := s.uploadImages(product.ID, images); err != nil {
			return nil, apperrors.New(
				http.StatusInternalServerError,
				"upload_product_images_failed",
				"failed to upload product images",
				err,
			)
		}
	}

	return mapper.ToProductResponse(product), nil
}

func (s *ProductService) DeleteProduct(merchantID uint, id uint) error {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(
				http.StatusNotFound,
				"product_not_found",
				"product not found",
				err,
			)
		}

		return apperrors.New(
			http.StatusInternalServerError,
			"fetch_product_failed",
			"failed to fetch product",
			err,
		)
	}

	if product.MerchantID != merchantID {
		return apperrors.New(
			http.StatusForbidden,
			"forbidden",
			"you do not own this product",
			nil,
		)
	}

	if err := s.productRepo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(
				http.StatusNotFound,
				"product_not_found",
				"product not found",
				err,
			)
		}

		return apperrors.New(
			http.StatusInternalServerError,
			"delete_product_failed",
			"failed to delete product",
			err,
		)
	}

	return nil
}

func (s *ProductService) GetMerchantProduct(
	merchantID uint,
	productID uint,
) (*dto.ProductResponse, error) {

	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(
				http.StatusNotFound,
				"product_not_found",
				"product not found",
				err,
			)
		}

		return nil, apperrors.New(
			http.StatusInternalServerError,
			"fetch_product_failed",
			"failed to fetch product",
			err,
		)
	}

	if product.MerchantID != merchantID {
		return nil, apperrors.New(
			http.StatusForbidden,
			"access_denied",
			"you do not have permission to access this product",
			nil,
		)
	}

	return mapper.ToProductResponse(product), nil
}
