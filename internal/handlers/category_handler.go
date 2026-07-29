package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"gocart/internal/dto"
	apperrors "gocart/internal/errors"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// CreateCategory godoc
//
//	@Summary		Create category
//	@Description	Create a new product category (Admin only)
//	@Tags			Categories
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CategoryRequest	true	"Category request"
//	@Success		201		{object}	dto.CategoryResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		403		{object}	errors.ErrorResponse
//	@Failure		409		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/admin/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {

	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetCategories godoc
//
//	@Summary		List categories
//	@Description	Get all product categories
//	@Tags			Categories
//	@Produce		json
//	@Success		200	{array}		dto.CategoryResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryByID godoc
//
//	@Summary		Get category
//	@Description	Get a category by ID
//	@Tags			Categories
//	@Produce		json
//	@Param			id	path		int	true	"Category ID"
//	@Success		200	{object}	dto.CategoryResponse
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_category_id",
			"invalid category id",
			err,
		))
		return
	}

	category, err := h.categoryService.GetCategoryByID(uint(id))
	if err != nil {
		if errors.Is(err, services.ErrCategoryNotFound) {
			c.Error(err)
			return
		}
	}

	c.JSON(http.StatusOK, category)
}

// UpdateCategory godoc
//
//	@Summary		Update category
//	@Description	Update an existing category (Admin only)
//	@Tags			Categories
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"Category ID"
//	@Param			request	body		dto.UpdateCategoryRequest	true	"Update category request"
//	@Success		200		{object}	dto.CategoryResponse
//	@Failure		400		{object}	errors.ErrorResponse
//	@Failure		401		{object}	errors.ErrorResponse
//	@Failure		403		{object}	errors.ErrorResponse
//	@Failure		404		{object}	errors.ErrorResponse
//	@Failure		409		{object}	errors.ErrorResponse
//	@Failure		500		{object}	errors.ErrorResponse
//	@Router			/api/v1/admin/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_category_id",
			"invalid category id",
			err,
		))
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"validation_error",
			err.Error(),
			err,
		))
		return
	}

	category, err := h.categoryService.UpdateCategory(&req, uint(id))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
//
//	@Summary		Delete category
//	@Description	Delete a category by ID (Admin only)
//	@Tags			Categories
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Category ID"
//	@Success		204
//	@Failure		400	{object}	errors.ErrorResponse
//	@Failure		401	{object}	errors.ErrorResponse
//	@Failure		403	{object}	errors.ErrorResponse
//	@Failure		404	{object}	errors.ErrorResponse
//	@Failure		500	{object}	errors.ErrorResponse
//	@Router			/api/v1/admin/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		c.Error(apperrors.New(
			http.StatusBadRequest,
			"invalid_category_id",
			"invalid category id",
			err,
		))
		return
	}

	if err := h.categoryService.DeleteCategory(uint(id)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
