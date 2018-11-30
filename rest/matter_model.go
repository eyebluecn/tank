package rest

/**
 * 文件。alien表示是否是应用内使用的文件，比如说蓝眼云盘的头像，alien = true 这种文件在上传时不需要指定存放目录，会统一放在同一个文件夹下。
 */
type Matter struct {
	Base
	Puuid    string  `json:"puuid"`
	UserUuid string  `json:"userUuid"`
	Dir      bool    `json:"dir"`
	Alien    bool    `json:"alien"`
	Name     string  `json:"name"`
	Md5      string  `json:"md5"`
	Size     int64   `json:"size"`
	Privacy  bool    `json:"privacy"`
	Path     string  `json:"path"`
	Times    int64   `json:"times"`
	Parent   *Matter `gorm:"-" json:"parent"`
}

// set File's table name to be `profiles`
func (Matter) TableName() string {
	return TABLE_PREFIX + "matter"
}
