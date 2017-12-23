package rest

import (
	"time"
)

type DownloadToken struct {
	Base
	UserUuid   string    `json:"userUuid"`
	MatterUuid string    `json:"matterUuid"`
	ExpireTime time.Time `json:"expireTime"`
	Ip         string    `json:"ip"`
}

func (DownloadToken) TableName() string {
	return TABLE_PREFIX + "download_token"
}
