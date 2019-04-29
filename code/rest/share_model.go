package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"time"
)

const (
	//单文件
	SHARE_TYPE_FILE = "FILE"
	//文件夹
	SHARE_TYPE_DIRECTORY = "DIRECTORY"
	//混合体
	SHARE_TYPE_MIX = "MIX"
)

const (
	SHARE_MAX_NUM = 100
)

/**
 * 分享记录
 */
type Share struct {
	Base
	Name           string    `json:"name" gorm:"type:varchar(255)"`
	ShareType      string    `json:"shareType" gorm:"type:varchar(45)"`
	Username       string    `json:"username" gorm:"type:varchar(45)"`
	UserUuid       string    `json:"userUuid" gorm:"type:char(36)"`
	DownloadTimes  int64     `json:"downloadTimes" gorm:"type:bigint(20) not null;default:0"`
	Code           string    `json:"code" gorm:"type:varchar(45) not null"`
	ExpireInfinity bool      `json:"expireInfinity" gorm:"type:tinyint(1) not null;default:0"`
	ExpireTime     time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	DirMatter      *Matter   `json:"dirMatter" gorm:"-"`
	Matters        []*Matter `json:"matters" gorm:"-"`
}

// set File's table name to be `profiles`
func (this *Share) TableName() string {
	return core.TABLE_PREFIX + "share"
}
