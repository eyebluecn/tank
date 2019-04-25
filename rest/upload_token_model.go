package rest

import (
	"tank/rest/config"
	"time"
)

type UploadToken struct {
	Base
	UserUuid   string    `json:"userUuid"`
	FolderUuid string    `json:"folderUuid"`
	MatterUuid string    `json:"matterUuid"`
	ExpireTime time.Time `json:"expireTime"`
	Filename   string    `json:"filename"`
	Privacy    bool      `json:"privacy"`
	Size       int64     `json:"size"`
	Ip         string    `json:"ip"`
}

func (this *UploadToken) TableName() string {
	return config.TABLE_PREFIX + "upload_token"
}
