package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"time"
)

type UploadToken struct {
	Base
	UserUuid   string    `json:"userUuid" gorm:"type:char(36) not null"`
	FolderUuid string    `json:"folderUuid" gorm:"type:char(36) not null"`
	MatterUuid string    `json:"matterUuid" gorm:"type:char(36) not null"`
	ExpireTime time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Filename   string    `json:"filename" gorm:"type:varchar(255) not null"`
	Privacy    bool      `json:"privacy" gorm:"type:tinyint(1) not null;default:0"`
	Size       int64     `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Ip         string    `json:"ip" gorm:"type:varchar(128) not null"`
}

func (this *UploadToken) TableName() string {
	return core.TABLE_PREFIX + "upload_token"
}
