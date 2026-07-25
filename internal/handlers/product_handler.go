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

func (h *ProductHandler) GetProducts(c *gin.Context) {

	q, f := query.NewProductQueryFromGin(c)

	resp, err := h.productService.GetProducts(q, f)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

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
