package handlers

import (
	"errors"
	"net/http"
	"strconv"

	apperrors "gocart/internal/errors"
	"gocart/internal/models"
	"gocart/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {

	var req models.CategoryRequest
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

func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, categories)
}

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

	var req models.UpdateCategoryRequest
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
