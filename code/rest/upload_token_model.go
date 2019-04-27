package rest

import (
	"github.com/eyebluecn/tank/code/core"
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
	return core.TABLE_PREFIX + "upload_token"
}
