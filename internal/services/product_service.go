package services

import (
	"errors"
	"fmt"
	"gocart/internal/models"
	"gocart/internal/repositories"
	"gocart/internal/storage"
	"mime/multipart"
)

var (
	ErrProductNotFound = errors.New("product not found")
	// ErrCategoryNotFound = errors.New("invalid category")
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
	req *models.CreateProductRequest,
	images []*multipart.FileHeader,
) (*models.Product, error) {

	_, err := s.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		Sku:         req.Sku,
		Slug:        req.Slug,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	if len(images) == 0 {
		if err := s.uploadImages(product.ID, images); err != nil {
			return nil, err
		}
	}

	return s.productRepo.GetByID(product.ID)
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

func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	return product, nil
}

func (s *ProductService) GetProducts(query *models.PaginationQuery, filters *models.ProductFilters) (*models.PaginatedResponse, error) {

	if query == nil {
		query = &models.PaginationQuery{
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
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:      products,
		Total:     total,
		Page:      query.Page,
		PageSize:  query.PageSize,
		TotalPage: totalPages,
	}, nil
}

func (s *ProductService) UpdateProduct(id uint, req *models.UpdateProductRequest) (*models.Product, error) {
	product, err := s.productRepo.GetByID(id)

	if err != nil {
		return nil, ErrProductNotFound
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

	if err := s.productRepo.Update((product)); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (s *ProductService) DeleteProduct(id uint) error {
	if err := s.productRepo.Delete(id); err != nil {
		return ErrProductNotFound
	}

	return nil
}
