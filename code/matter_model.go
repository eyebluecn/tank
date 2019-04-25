package code

import (
	"fmt"
	"tank/code/config"
	"tank/code/tool"
)

const (
	//根目录的uuid
	MATTER_ROOT  = "root"
	//cache文件夹名称
	MATTER_CACHE = "cache"
	//matter名称最大长度
	MATTER_NAME_MAX_LENGTH = 200
	//matter文件夹最大深度
	MATTER_NAME_MAX_DEPTH = 32

)

/**
 * 文件。
 */
type Matter struct {
	Base
	Puuid    string  `json:"puuid" gorm:"type:char(36);index:idx_puuid"`
	UserUuid string  `json:"userUuid" gorm:"type:char(36);index:idx_uu"`
	Username string  `json:"username" gorm:"type:varchar(45) not null"`
	Dir      bool    `json:"dir" gorm:"type:tinyint(1) not null;default:0"`
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
	return config.TABLE_PREFIX + "matter"
}

// 获取该Matter的绝对路径。path代表的是相对路径。
func (this *Matter) AbsolutePath() string {
	return GetUserFileRootDir(this.Username) + this.Path
}

// 获取该Matter的MimeType
func (this *Matter) MimeType() string {
	return tool.GetMimeType(tool.GetExtension(this.Name))
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

//获取到用户文件的根目录。
func GetUserFileRootDir(username string) (rootDirPath string) {

	rootDirPath = fmt.Sprintf("%s/%s/%s", config.CONFIG.MatterPath, username, MATTER_ROOT)

	return rootDirPath
}

//获取到用户缓存的根目录。
func GetUserCacheRootDir(username string) (rootDirPath string) {

	rootDirPath = fmt.Sprintf("%s/%s/%s", config.CONFIG.MatterPath, username, MATTER_CACHE)

	return rootDirPath
}

