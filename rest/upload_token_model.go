package rest

import (
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

func (UploadToken) TableName() string {
	return TABLE_PREFIX + "upload_token"
}
