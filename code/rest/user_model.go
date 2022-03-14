package rest

import (
	"time"
)

const (
	//guest
	USER_ROLE_GUEST = "GUEST"
	//normal user
	USER_ROLE_USER = "USER"
	//administrator
	USER_ROLE_ADMINISTRATOR = "ADMINISTRATOR"
)

const (
	//ok
	USER_STATUS_OK = "OK"
	//disabled
	USER_STATUS_DISABLED = "DISABLED"
)

const (
	//username pattern
	USERNAME_PATTERN = "^[\\p{Han}0-9a-zA-Z_]+$"
	USERNAME_DEMO    = "demo"
)

type User struct {
	Uuid           string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort           int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime     time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime     time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Role           string    `json:"role" gorm:"type:varchar(45)"`
	Username       string    `json:"username" gorm:"type:varchar(45) not null;unique"`
	Password       string    `json:"-" gorm:"type:varchar(255)"`
	AvatarUrl      string    `json:"avatarUrl" gorm:"type:varchar(255)"`
	LastIp         string    `json:"lastIp" gorm:"type:varchar(128)"`
	LastTime       time.Time `json:"lastTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	SizeLimit      int64     `json:"sizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	TotalSizeLimit int64     `json:"totalSizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	TotalSize      int64     `json:"totalSize" gorm:"type:bigint(20) not null;default:0"`
	Status         string    `json:"status" gorm:"type:varchar(45)"`
}
