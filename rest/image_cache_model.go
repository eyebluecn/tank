package rest

/**
 * 图片缓存，对于那些处理过的图片，统一管理在这里。
 */
type ImageCache struct {
	Base
	UserUuid   string  `json:"userUuid" gorm:"type:char(36)"`
	MatterUuid string  `json:"matterUuid" gorm:"type:char(36);index:idx_mu"`
	Mode       string  `json:"mode" gorm:"type:varchar(512)"`
	Md5        string  `json:"md5" gorm:"type:varchar(45)"`
	Size       int64   `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Path       string  `json:"path" gorm:"type:varchar(512)"`
	Matter     *Matter `json:"matter" gorm:"-"`
}

// set File's table name to be `profiles`
func (ImageCache) TableName() string {
	return TABLE_PREFIX + "image_cache"
}
