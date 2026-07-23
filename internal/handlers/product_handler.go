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

	var images []*multipart.FileHeader

	form, err := c.MultipartForm()
	if err == nil {
		images = form.File["images"]
	}

	product, err := h.productService.CreateProduct(&req, images)
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

	product, err := h.productService.UpdateProduct(uint(id), &req, images)
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

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
