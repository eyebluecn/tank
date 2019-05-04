package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

/**
 * the link table for Share and Matter.
 */
type Bridge struct {
	Base
	ShareUuid  string `json:"shareUuid" gorm:"type:char(36)"`
	MatterUuid string `json:"matterUuid" gorm:"type:char(36)"`
}

func (this *Bridge) TableName() string {
	return core.TABLE_PREFIX + "bridge"
}
