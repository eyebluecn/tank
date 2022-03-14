package rest

import "time"

/**
 * the link table for Share and Matter.
 */
type Bridge struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	ShareUuid  string    `json:"shareUuid" gorm:"type:char(36)"`
	MatterUuid string    `json:"matterUuid" gorm:"type:char(36)"`
}
