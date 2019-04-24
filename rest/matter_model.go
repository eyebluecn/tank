package rest

import "tank/rest/util"

const (
	MATTER_ROOT  = "root"
	MATTER_CACHE = "cache"
)

/**
 * 文件。alien表示是否是应用内使用的文件，比如说蓝眼云盘的头像，alien = true 这种文件在上传时不需要指定存放目录，会统一放在同一个文件夹下。
 */
type Matter struct {
	Base
	Puuid    string  `json:"puuid" gorm:"type:char(36);index:idx_puuid"`
	UserUuid string  `json:"userUuid" gorm:"type:char(36);index:idx_uu"`
	Username string  `json:"username" gorm:"type:varchar(45) not null"`
	Dir      bool    `json:"dir" gorm:"type:tinyint(1) not null;default:0"`
	Alien    bool    `json:"alien" gorm:"type:tinyint(1) not null;default:0"`
	Name     string  `json:"name" gorm:"type:varchar(255) not null"`
	Md5      string  `json:"md5" gorm:"type:varchar(45)"`
	Size     int64   `json:"size" gorm:"type:bigint(20) not null;default:0"`
	Privacy  bool    `json:"privacy" gorm:"type:tinyint(1) not null;default:0"`
	Path     string  `json:"path" gorm:"type:varchar(512)"`
	Times    int64   `json:"times" gorm:"type:bigint(20) not null;default:0"`
	Parent   *Matter `json:"parent" gorm:"-"`
}

// set File's table name to be `profiles`
func (Matter) TableName() string {
	return TABLE_PREFIX + "matter"
}

// 获取该Matter的绝对路径。path代表的是相对路径。
func (this *Matter) AbsolutePath() string {
	return GetUserFileRootDir(this.Username) + this.Path
}

// 获取该Matter的MimeType
func (this *Matter) MimeType() string {
	return util.GetMimeType(util.GetExtension(this.Name))
}


//创建一个 ROOT 的matter，主要用于统一化处理移动复制等内容。
func NewRootMatter(user *User) *Matter {
	matter := &Matter{}
	matter.Uuid = MATTER_ROOT
	matter.UserUuid = user.Uuid
	matter.Username = user.Username
	matter.Dir = true
	matter.Path = ""
	matter.CreateTime = user.CreateTime
	matter.UpdateTime = user.UpdateTime

	return matter
}
