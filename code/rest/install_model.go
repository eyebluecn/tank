package rest

import "github.com/jinzhu/gorm"

/**
 * 表名对应的表结构
 */
type InstallTableInfo struct {
	Name          string              `json:"name"`
	TableExist    bool                `json:"tableExist"`
	AllFields     []*gorm.StructField `json:"allFields"`
	MissingFields []*gorm.StructField `json:"missingFields"`
}
