package rest

import "github.com/jinzhu/gorm"

/**
 * table meta info.
 */
type InstallTableInfo struct {
	Name          string              `json:"name"`
	TableExist    bool                `json:"tableExist"`
	AllFields     []*gorm.StructField `json:"allFields"`
	MissingFields []*gorm.StructField `json:"missingFields"`
}
