package rest

import (
	"math"
	"time"
)

const (
	TRUE  = "true"
	FALSE = "false"

	DIRECTION_ASC  = "ASC"
	DIRECTION_DESC = "DESC"
)

type IBase interface {
	//name of db table
	TableName() string
}

// Mysql 5.5 only support one CURRENT_TIMESTAMP
// so we use 2018-01-01 00:00:00 as default, which is the first release date of EyeblueTank
type Base struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
}

func (this *Base) TableName() string {
	panic("you should overwrite TableName()")
}

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
