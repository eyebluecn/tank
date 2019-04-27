package rest

import "github.com/eyebluecn/tank/code/core"

type Preference struct {
	Base
	Name        string `json:"name" gorm:"type:varchar(45)"`
	LogoUrl     string `json:"logoUrl" gorm:"type:varchar(255)"`
	FaviconUrl  string `json:"faviconUrl" gorm:"type:varchar(255)"`
	FooterLine1 string `json:"footerLine1" gorm:"type:varchar(1024)"`
	FooterLine2 string `json:"footerLine2" gorm:"type:varchar(1024)"`
	ShowAlien   bool   `json:"showAlien" gorm:"type:tinyint(1) not null;default:1"`
}

// set File's table name to be `profiles`
func (this *Preference) TableName() string {
	return core.TABLE_PREFIX + "preference"
}
