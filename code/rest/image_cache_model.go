package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

/**
 * image cache.
 */
type ImageCache struct {
	Base
	Name       string  `json:"name" gorm:"type:varchar(255) not null"`
	UserUuid   string  `json:"userUuid" gorm:"type:char(36)"`
	Username   string  `json:"username" gorm:"type:varchar(45) not null"`
	MatterUuid string  `json:"matterUuid" gorm:"type:char(36);index:idx_mu"`
	MatterName string  `json:"matterName" gorm:"type:varchar(255) not null"`
	Mode       string  `json:"mode" gorm:"type:varchar(512)"`
	Md5        string  `json:"md5" gorm:"type:varchar(45)"`
	Size       int64   `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Path       string  `json:"path" gorm:"type:varchar(512)"`
	Matter     *Matter `json:"matter" gorm:"-"`
}

// set File's table name to be `profiles`
func (this *ImageCache) TableName() string {
	return core.TABLE_PREFIX + "image_cache"
}

// get the absolute path. path in db means relative path.
func (this *ImageCache) AbsolutePath() string {
	return GetUserCacheRootDir(this.Username) + this.Path
}
