package query

import (
	"strconv"

	"gocart/internal/models"

	"github.com/gin-gonic/gin"
)

func NewProductQueryFromGin(c *gin.Context) (*models.PaginationQuery, *models.ProductFilters) {
	query := &models.PaginationQuery{
		Page:     1,
		PageSize: 10,
		Sort:     "created_at",
		Order:    "desc",
	}

	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			query.Page = p
		}
	}

	if v := c.Query("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 && ps <= 100 {
			query.PageSize = ps
		}
	}

	if v := c.Query("sort"); v != "" {
		query.Sort = v
	}

	if v := c.Query("order"); v == "asc" || v == "desc" {
		query.Order = v
	}

	filters := &models.ProductFilters{}

	if v := c.Query("category_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			filters.CategoryID = uint(id)
		}
	}

	if v := c.Query("min_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			filters.MinPrice = p
		}
	}

	if v := c.Query("max_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			filters.MaxPrice = p
		}
	}

	if v := c.Query("search"); v != "" {
		filters.SearchQuery = v
	}

	if v := c.Query("in_stock"); v == "true" {
		b := true
		filters.InStock = &b
	}

	return query, filters
}
