package models

type PaginationQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Sort     string `form:"sort,default=created_at"`
	Order    string `form:"order,default=desc"`
}

type PaginatedResponse struct {
	Data      interface{} `json:"data"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	TotalPage int         `json:"total_pages"`
}
