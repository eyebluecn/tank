package rest

import (
	"time"
)

const (
	//single file.
	SHARE_TYPE_FILE = "FILE"
	//directory
	SHARE_TYPE_DIRECTORY = "DIRECTORY"
	//mix things
	SHARE_TYPE_MIX = "MIX"
)

const (
	SHARE_MAX_NUM = 100
)

/**
 * share record
 */
type Share struct {
	Uuid           string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort           int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime     time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime     time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Name           string    `json:"name" gorm:"type:varchar(255)"`
	ShareType      string    `json:"shareType" gorm:"type:varchar(45)"`
	Username       string    `json:"username" gorm:"type:varchar(45)"`
	UserUuid       string    `json:"userUuid" gorm:"type:char(36)"`
	DownloadTimes  int64     `json:"downloadTimes" gorm:"type:bigint(20) not null;default:0"`
	Code           string    `json:"code" gorm:"type:varchar(45) not null"`
	ExpireInfinity bool      `json:"expireInfinity" gorm:"type:tinyint(1) not null;default:0"`
	ExpireTime     time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	SpaceUuid      string    `json:"spaceUuid" gorm:"type:char(36) not null;index:idx_share_space_uuid"`
	DirMatter      *Matter   `json:"dirMatter" gorm:"-"`
	Matters        []*Matter `json:"matters" gorm:"-"`
}
