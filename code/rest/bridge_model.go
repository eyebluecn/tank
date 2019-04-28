package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

/**
 * 分享记录和matter的关联表
 */
type Bridge struct {
	Base
	ShareUuid  string `json:"shareUuid" gorm:"type:char(36)"`
	MatterUuid string `json:"matterUuid" gorm:"type:char(36)"`
}

// set File's table name to be `profiles`
func (this *Bridge) TableName() string {
	return core.TABLE_PREFIX + "bridge"
}
