package rest

import (
	jsoniter "github.com/json-iterator/go"
	"time"
)

type Preference struct {
	Uuid                  string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort                  int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime            time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime            time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	Name                  string    `json:"name" gorm:"type:varchar(45)"`
	LogoUrl               string    `json:"logoUrl" gorm:"type:varchar(255)"`
	FaviconUrl            string    `json:"faviconUrl" gorm:"type:varchar(255)"`
	Copyright             string    `json:"copyright" gorm:"type:varchar(1024)"`
	Record                string    `json:"record" gorm:"type:varchar(1024)"`
	DownloadDirMaxSize    int64     `json:"downloadDirMaxSize" gorm:"type:bigint(20) not null;default:-1"`
	DownloadDirMaxNum     int64     `json:"downloadDirMaxNum" gorm:"type:bigint(20) not null;default:-1"`
	DefaultTotalSizeLimit int64     `json:"defaultTotalSizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	AllowRegister         bool      `json:"allowRegister" gorm:"type:tinyint(1) not null;default:0"`
	PreviewConfig         string    `json:"previewConfig" gorm:"type:text"`
	ScanConfig            string    `json:"scanConfig" gorm:"type:text"`
	DeletedKeepDays       int64     `json:"deletedKeepDays" gorm:"type:bigint(20) not null;default:7"`
	Version               string    `json:"version" gorm:"-"`
}

const (
	//scan scope all.
	SCAN_SCOPE_ALL = "ALL"
	//scan scope custom.
	SCAN_SCOPE_CUSTOM = "CUSTOM"
)

//scan config struct.
type ScanConfig struct {
	//whether enable the scan task.
	Enable bool `json:"enable"`
	//when to process the task. five fields. @every 1s
	Cron string `json:"cron"`
	//username
	Usernames []string `json:"usernames"`
	//scan scope. see SCAN_SCOPE
	Scope string `json:"scope"`
}

//fetch the scan config
func (this *Preference) FetchScanConfig() *ScanConfig {

	json := this.ScanConfig
	if json == "" || json == EMPTY_JSON_MAP {

		return &ScanConfig{
			Enable: false,
		}
	} else {
		m := &ScanConfig{}

		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(json), &m)
		if err != nil {
			panic(err)
		}
		return m
	}
}
