package rest

/**
 * 图片缓存，对于那些处理过的图片，统一管理在这里。
 */
type ImageCache struct {
	Base
	UserUuid   string  `json:"userUuid"`
	MatterUuid string  `json:"matterUuid"`
	Mode       string  `json:"mode"`
	Md5        string  `json:"md5"`
	Size       int64   `json:"size"`
	Path       string  `json:"path"`
	Matter     *Matter `gorm:"-" json:"matter"`
}

// set File's table name to be `profiles`
func (ImageCache) TableName() string {
	return TABLE_PREFIX + "image_cache"
}
