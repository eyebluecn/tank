package rest

import (
	"math"
	"reflect"
	"time"
)

type Time time.Time

type IBase interface {
	//返回其对应的数据库表名
	TableName() string
}

type Base struct {
	Uuid       string    `gorm:"primary_key" json:"uuid"`
	Sort       int64     `json:"sort"`
	UpdateTime time.Time `json:"updateTime"`
	CreateTime time.Time `json:"createTime"`
}

//将 Struct 转换成map[string]interface{}类型
func (this *Base) Map() map[string]interface{} {
	t := reflect.TypeOf(this)
	v := reflect.ValueOf(this)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

func (Base) TableName() string {
	return TABLE_PREFIX + "base"
}

//分页类
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
