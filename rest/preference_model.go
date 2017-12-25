package rest

type Preference struct {
	Base
	Name        string `json:"name"`
	LogoUrl     string `json:"logoUrl"`
	FaviconUrl  string `json:"faviconUrl"`
	FooterLine1 string `json:"footerLine1"`
	FooterLine2 string `json:"footerLine2"`
}

// set File's table name to be `profiles`
func (Preference) TableName() string {
	return TABLE_PREFIX + "preference"
}
