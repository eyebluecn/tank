package rest

import (
	"math"
)

const (
	TRUE  = "true"
	FALSE = "false"

	DIRECTION_ASC  = "ASC"
	DIRECTION_DESC = "DESC"

	EMPTY_JSON_MAP   = "{}"
	EMPTY_JSON_ARRAY = "[]"
)

//pager
type Pager struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalItems int         `json:"totalItems"`
	TotalPages int         `json:"totalPages"`
	Data       interface{} `json:"data"`
}

func NewPager(page int, pageSize int, totalItems int, data interface{}) *Pager {

	return &Pager{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: int(math.Ceil(float64(totalItems) / float64(pageSize))),
		Data:       data,
	}
}
