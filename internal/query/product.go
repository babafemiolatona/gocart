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
	query := ParsePagination(c)

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
