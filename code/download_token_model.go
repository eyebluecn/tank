package code

import (
	"tank/code/config"
	"time"
)

type DownloadToken struct {
	Base
	UserUuid   string    `json:"userUuid" gorm:"type:char(36) not null"`
	MatterUuid string    `json:"matterUuid" gorm:"type:char(36) not null;index:idx_mu"`
	ExpireTime time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Ip         string    `json:"ip" gorm:"type:varchar(128) not null"`
}

func (this *DownloadToken) TableName() string {
	return config.TABLE_PREFIX + "download_token"
}
