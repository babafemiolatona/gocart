package query

import (
	"strconv"

	"gocart/internal/dto"

	"github.com/gin-gonic/gin"
)

func ParsePagination(c *gin.Context) *dto.PaginationQuery {
	q := &dto.PaginationQuery{
		Page:     1,
		PageSize: 10,
		Sort:     "created_at",
		Order:    "desc",
	}

	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			q.Page = p
		}
	}

	if v := c.Query("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 && ps <= 100 {
			q.PageSize = ps
		}
	}

	if v := c.Query("sort"); v != "" {
		q.Sort = v
	}

	if v := c.Query("order"); v == "asc" || v == "desc" {
		q.Order = v
	}

	return q
}
