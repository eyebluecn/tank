package rest

import (
	"time"
)

const (
	//private space
	SPACE_TYPE_PRIVATE = "PRIVATE"
	//group shared space.
	SPACE_TYPE_SHARED = "SHARED"
)

/**
 * shared space
 */
type Space struct {
	Uuid           string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort           int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime     time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime     time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Name           string    `json:"name" gorm:"type:varchar(100) not null;unique"`
	UserUuid       string    `json:"userUuid" gorm:"type:char(36)"`
	SizeLimit      int64     `json:"sizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	TotalSizeLimit int64     `json:"totalSizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	TotalSize      int64     `json:"totalSize" gorm:"type:bigint(20) not null;default:0"`
	Type           string    `json:"type" gorm:"type:varchar(45)"`
	User           *User     `json:"user" gorm:"-"`
}
