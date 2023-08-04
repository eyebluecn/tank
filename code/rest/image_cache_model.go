package rest

import "time"

/**
 * image cache.
 */
type ImageCache struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Name       string    `json:"name" gorm:"type:varchar(255) not null"`
	UserUuid   string    `json:"userUuid" gorm:"type:char(36)"`
	Username   string    `json:"username" gorm:"type:varchar(45) not null"`
	MatterUuid string    `json:"matterUuid" gorm:"type:char(36);index:idx_image_cache_mu"` //index should unique globally.
	MatterName string    `json:"matterName" gorm:"type:varchar(255) not null"`
	Mode       string    `json:"mode" gorm:"type:varchar(512)"`
	Md5        string    `json:"md5" gorm:"type:varchar(45)"`
	Size       int64     `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Path       string    `json:"path" gorm:"type:varchar(512)"`
	Matter     *Matter   `json:"matter" gorm:"-"`
}

// get the absolute path. path in db means relative path.
func (this *ImageCache) AbsolutePath() string {
	return GetSpaceCacheRootDir(this.Username) + this.Path
}
