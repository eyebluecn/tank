package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"time"
)

/**
 * 分享记录
 */
type Share struct {
	Base
	UserUuid      string    `json:"userUuid" gorm:"type:char(36)"`
	DownloadTimes int64     `json:"downloadTimes" gorm:"type:bigint(20) not null;default:0"`
	Code          string    `json:"code" gorm:"type:varchar(45) not null"`
	ExpireTime    time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
}

// set File's table name to be `profiles`
func (this *Share) TableName() string {
	return core.TABLE_PREFIX + "share"
}
