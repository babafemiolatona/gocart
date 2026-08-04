package query

import (
	"strconv"
	"strings"

	"gocart/internal/dto"

	"github.com/gin-gonic/gin"
)

type ProductFilters struct {
	CategoryID  uint
	MerchantID  uint
	MinPrice    int64
	MaxPrice    int64
	InStock     *bool
	SearchQuery string
}

func NewProductQueryFromGin(c *gin.Context) (*dto.PaginationQuery, *ProductFilters) {
	query := &dto.PaginationQuery{
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

	filters := &ProductFilters{}

	if v := c.Query("category_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 32); err == nil {
			filters.CategoryID = uint(id)
		}
	}

	if v := c.Query("min_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			filters.MinPrice = int64(p * 100)
		}
	}

	if v := c.Query("max_price"); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			filters.MaxPrice = int64(p * 100)
		}
	}

	if v := strings.TrimSpace(c.Query("search")); v != "" {
		filters.SearchQuery = v
	}

	if v := c.Query("in_stock"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filters.InStock = &b
		}
	}

	return query, filters
}
