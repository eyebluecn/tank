package rest

import (
	"time"
)

type UploadToken struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	UserUuid   string    `json:"userUuid" gorm:"type:char(36) not null"`
	FolderUuid string    `json:"folderUuid" gorm:"type:char(36) not null"`
	MatterUuid string    `json:"matterUuid" gorm:"type:char(36) not null"`
	ExpireTime time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Filename   string    `json:"filename" gorm:"type:varchar(255) not null"`
	Privacy    bool      `json:"privacy" gorm:"type:tinyint(1) not null;default:0"`
	Size       int64     `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Ip         string    `json:"ip" gorm:"type:varchar(128) not null"`
}
