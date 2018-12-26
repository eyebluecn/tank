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

//Mysql 5.5只支持一个CURRENT_TIMESTAMP的默认值，因此时间的默认值都使用蓝眼云盘第一个发布版本时间 2018-01-01 00:00:00
type Base struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
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
