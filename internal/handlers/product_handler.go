package handlers

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/query"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// CreateProduct godoc
//
//	@Summary		Create product
//	@Description	Create a new product (Merchant only)
//	@Tags			Products
//	@Security		BearerAuth
//	@Accept			mpfd
//	@Produce		json
//	@Param			name			formData	string	true	"Product name"
//	@Param			description		formData	string	false	"Product description"
//	@Param			price			formData	number	true	"Product price"
//	@Param			stock			formData	int		true	"Stock quantity"
//	@Param			category_id		formData	int		true	"Category ID"
//	@Param			sku				formData	string	true	"Stock Keeping Unit (SKU)"
//	@Param			slug			formData	string	true	"Product slug"
//	@Param			images			formData	file	false	"Product images"
//	@Success		201				{object}	dto.ProductResponse
//	@Failure		400				{object}	errors.ErrorResponse
//	@Failure		401				{object}	errors.ErrorResponse
//	@Failure		403				{object}	errors.ErrorResponse
//	@Failure		404				{object}	errors.ErrorResponse
//	@Failure		409				{object}	errors.ErrorResponse
//	@Failure		500				{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest

	if err := c.ShouldBind(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var images []*multipart.FileHeader

	form, err := c.MultipartForm()
	if err == nil {
		images = form.File["images"]
	}

	product, err := h.productService.CreateProduct(merchantID, &req, images)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProducts godoc
//
//	@Summary		List products
//	@Description	Get paginated list of products
//	@Tags			Products
//	@Produce		json
//	@Param			page			query		int		false	"Page number"
//	@Param			page_size		query		int		false	"Page size"
//	@Param			search			query		string	false	"Search by product name"
//	@Param			category_id	query		int		false	"Category ID"
//	@Param			min_price		query		number	false	"Minimum price"
//	@Param			max_price		query		number	false	"Maximum price"
//	@Param			in_stock		query		bool	false	"Only in-stock products"
//	@Success		200				{object}	dto.PaginatedResponse
//	@Failure		500				{object}	errors.ErrorResponse
//	@Router			/api/v1/products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {

	q, f := query.NewProductQueryFromGin(c)

	resp, err := h.productService.GetProducts(q, f)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProduct godoc
//
//	@Summary		Get product
//	@Description	Get a product by ID
//	@Tags			Products
//	@Produce		json
//	@Param			id	path		int	true	"Product ID"
//	@Success		200	{object}	dto.ProductResponse
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_product_id",
			"invalid product id",
			err,
		))
		return
	}

	product, err := h.productService.GetProduct(uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct godoc
//
//	@Summary		Update product
//	@Description	Update an existing product (Merchant only)
//	@Tags			Products
//	@Security		BearerAuth
//	@Accept			mpfd
//	@Produce		json
//	@Param			id				path		int		true	"Product ID"
//	@Param			name			formData	string	false	"Product name"
//	@Param			description	formData	string	false	"Product description"
//	@Param			price			formData	number	false	"Product price"
//	@Param			stock			formData	int		false	"Stock quantity"
//	@Param			category_id	formData	int		false	"Category ID"
//	@Param			images			formData	file	false	"Product images"
//	@Success		200				{object}	dto.ProductResponse
//	@Failure		400				{object}	errors.ErrorResponse
//	@Failure		401				{object}	errors.ErrorResponse
//	@Failure		403				{object}	errors.ErrorResponse
//	@Failure		404				{object}	errors.ErrorResponse
//	@Failure		409				{object}	errors.ErrorResponse
//	@Failure		500				{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_product_id",
			"invalid product id",
			err,
		))
		return
	}

	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	var images []*multipart.FileHeader

	form, err := c.MultipartForm()
	if err == nil {
		images = form.File["images"]
	}

	product, err := h.productService.UpdateProduct(merchantID, uint(id), &req, images)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
//
//	@Summary		Delete product
//	@Description	Delete a product (Merchant only)
//	@Tags			Products
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Product ID"
//	@Success		204
//	@Failure		400				{object}	errors.ErrorResponse
//	@Failure		401				{object}	errors.ErrorResponse
//	@Failure		403				{object}	errors.ErrorResponse
//	@Failure		404				{object}	errors.ErrorResponse
//	@Failure		500				{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_product_id",
			"invalid product id",
			err,
		))
		return
	}

	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.productService.DeleteProduct(merchantID, uint(id)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetMerchantProducts godoc
//
//	@Summary		List my products
//	@Description	Get all products belonging to the authenticated merchant
//	@Tags			Products
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page			query		int		false	"Page number"
//	@Param			page_size		query		int		false	"Page size"
//	@Param			search			query		string	false	"Search by product name"
//	@Param			category_id	query		int		false	"Category ID"
//	@Param			min_price		query		number	false	"Minimum price"
//	@Param			max_price		query		number	false	"Maximum price"
//	@Param			in_stock		query		bool	false	"Only in-stock products"
//	@Success		200				{object}	dto.PaginatedResponse
//	@Failure		401				{object}	errors.ErrorResponse
//	@Failure		403				{object}	errors.ErrorResponse
//	@Failure		500				{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/products [get]
func (h *ProductHandler) GetMerchantProducts(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	q, f := query.NewProductQueryFromGin(c)

	f.MerchantID = merchantID

	resp, err := h.productService.GetProducts(q, f)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMerchantProduct godoc
//
//	@Summary		Get my product
//	@Description	Get one product belonging to the authenticated merchant
//	@Tags			Products
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		int	true	"Product ID"
//	@Success		200	{object}	dto.ProductResponse
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/merchants/products/{id} [get]
func (h *ProductHandler) GetMerchantProduct(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.Error(err)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_product_id",
			"invalid product id",
			err,
		))
		return
	}

	product, err := h.productService.GetMerchantProduct(
		merchantID,
		uint(id),
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}
