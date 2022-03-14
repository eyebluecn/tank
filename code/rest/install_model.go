package rest

import (
	"gorm.io/gorm/schema"
)

/**
 * table meta info.
 */
type InstallTableInfo struct {
	Name          string          `json:"name"`
	TableExist    bool            `json:"tableExist"`
	AllFields     []*schema.Field `json:"allFields"`
	MissingFields []*schema.Field `json:"missingFields"`
}
