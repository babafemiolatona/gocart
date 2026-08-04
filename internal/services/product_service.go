package services

import (
	"errors"
	"fmt"
	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/logger"
	"gocart/internal/mapper"
	"gocart/internal/query"
	"net/http"
	"strings"

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
	maxUploadSize    int64
}

func NewProductService(
	productRepo repositories.ProductRepository,
	categoryRepo repositories.CategoryRepository,
	productImageRepo repositories.ProductImageRepository,
	storage storage.Storage,
	maxUploadSize int64,
) *ProductService {
	return &ProductService{
		productRepo:      productRepo,
		categoryRepo:     categoryRepo,
		productImageRepo: productImageRepo,
		storage:          storage,
		maxUploadSize:    maxUploadSize,
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
		Price:       mapper.UnitToMinorUnits(req.Price),
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
			return nil, err
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

		if image.Size > s.maxUploadSize {
			return apperrors.New(
				http.StatusBadRequest,
				"file_too_large",
				fmt.Sprintf("image exceeds maximum size of %d bytes", s.maxUploadSize),
				nil,
			)
		}

		if !strings.HasPrefix(image.Header.Get("Content-Type"), "image/") {
			return apperrors.New(
				http.StatusBadRequest,
				"invalid_image_type",
				"only image files are allowed",
				nil,
			)
		}

		file, err := image.Open()
		if err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"upload_product_images_failed",
				"failed to open image",
				err,
			)
		}

		objectName, err := s.storage.UploadProductImage(
			file,
			image,
			productID,
		)
		file.Close()

		if err != nil {
			return apperrors.New(
				http.StatusInternalServerError,
				"upload_product_images_failed",
				"failed to upload image",
				err,
			)
		}

		productImages = append(productImages, models.ProductImage{
			ProductID: productID,
			ImageURL:  objectName,
		})
	}

	if len(productImages) > 0 {
		if err := s.productImageRepo.CreateMany(productImages); err != nil {
			for _, image := range productImages {
				_ = s.storage.DeleteObject(image.ImageURL)
			}
			return apperrors.New(
				http.StatusInternalServerError,
				"upload_product_images_failed",
				"failed to save product images",
				err,
			)
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

	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = mapper.UnitToMinorUnits(*req.Price)
	}
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}
	if req.CategoryID != nil {
		if _, err := s.categoryRepo.GetByID(*req.CategoryID); err != nil {
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

		updates["category_id"] = *req.CategoryID
	}
	if req.Sku != nil {
		updates["sku"] = *req.Sku
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}

	if len(updates) > 0 {
		if err := s.productRepo.Update(id, updates); err != nil {
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
	}

	if len(images) > 0 {
		if err := s.uploadImages(id, images); err != nil {
			return nil, err
		}
	}

	product, err = s.productRepo.GetByID(id)
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

	for _, image := range product.Images {
		if err := s.storage.DeleteObject(image.ImageURL); err != nil {
			logger.Log.Warn().
				Uint("product_id", id).
				Str("object", image.ImageURL).
				Err(err).
				Msg("failed to delete product image from storage")
		}
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
