package rest

import (
	"fmt"
	"time"

	"github.com/eyebluecn/tank/code/core"
)

type MatterChunk struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	UserUuid   string    `json:"userUuid" gorm:"type:char(36);index:idx_matter_uu"`
	Username   string    `json:"username" gorm:"type:varchar(45) not null"`
	Md5        string    `json:"md5" gorm:"type:varchar(45)"`
	Index      string    `json:"sort" gorm:"type:bigint(20) not null"`
}

// get matterChunk's absolute path. the Path property is relative path in db.
func (this *MatterChunk) AbsolutePath() string {
	return GetUserMatterChunkDir(this.Username)
}

//get user's matterChunk absolute path
func GetUserMatterChunkDir(username string) (rootDirPath string) {
	// FIXME dirPath
	dirPath := fmt.Sprintf("%s/%s/%s", core.CONFIG.MatterPath(), username, MATTER_ROOT)

	return dirPath
}
