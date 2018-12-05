package rest

type Preference struct {
	Base
	Name        string `json:"name" gorm:"type:varchar(45)"`
	LogoUrl     string `json:"logoUrl" gorm:"type:varchar(255)"`
	FaviconUrl  string `json:"faviconUrl" gorm:"type:varchar(255)"`
	FooterLine1 string `json:"footerLine1" gorm:"type:varchar(1024)"`
	FooterLine2 string `json:"footerLine2" gorm:"type:varchar(1024)"`
	Version     string `json:"version" gorm:"type:varchar(45)"`
}

// set File's table name to be `profiles`
func (Preference) TableName() string {
	return TABLE_PREFIX + "preference"
}
