package util

import "strconv"

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"-"`
}

func ParsePagination(pageValue, pageSizeValue string, defaultSize, maxSize int) Pagination {
	page, _ := strconv.Atoi(pageValue)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeValue)
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	return Pagination{Page: page, PageSize: pageSize, Offset: (page - 1) * pageSize}
}
