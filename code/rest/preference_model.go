package rest

import "github.com/eyebluecn/tank/code/core"

type Preference struct {
	Base
	Name                  string `json:"name" gorm:"type:varchar(45)"`
	LogoUrl               string `json:"logoUrl" gorm:"type:varchar(255)"`
	FaviconUrl            string `json:"faviconUrl" gorm:"type:varchar(255)"`
	Copyright             string `json:"copyright" gorm:"type:varchar(1024)"`
	Record                string `json:"record" gorm:"type:varchar(1024)"`
	DownloadDirMaxSize    int64  `json:"downloadDirMaxSize" gorm:"type:bigint(20) not null;default:-1"`
	DownloadDirMaxNum     int64  `json:"downloadDirMaxNum" gorm:"type:bigint(20) not null;default:-1"`
	DefaultTotalSizeLimit int64  `json:"defaultTotalSizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	AllowRegister         bool   `json:"allowRegister" gorm:"type:tinyint(1) not null;default:0"`
	Version               string `json:"version" gorm:"-"`
}

// set File's table name to be `profiles`
func (this *Preference) TableName() string {
	return core.TABLE_PREFIX + "preference"
}
