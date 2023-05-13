package rest

import (
	"time"
)

/**
 * shared space
 */
type Space struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	UserUuid   string    `json:"userUuid" gorm:"type:char(36);unique:uk_user_uuid"`
	User       *User     `json:"user" gorm:"-"`
}
