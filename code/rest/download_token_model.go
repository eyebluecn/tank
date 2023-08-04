package rest

import (
	"time"
)

type DownloadToken struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	UserUuid   string    `json:"userUuid" gorm:"type:char(36) not null"`
	MatterUuid string    `json:"matterUuid" gorm:"type:char(36) not null;index:idx_download_token_mu"` //index should unique globally for sqlite
	ExpireTime time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Ip         string    `json:"ip" gorm:"type:varchar(128) not null"`
}
