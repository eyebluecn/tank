package rest

import "time"

// Footprint /**
// Mysql 5.5 only support one CURRENT_TIMESTAMP
// so we use 2018-01-01 00:00:00 as default, which is the first release date of EyeblueTank
type Footprint struct {
	Uuid       string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort       int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	UserUuid   string    `json:"userUuid" gorm:"type:char(36)"`
	Ip         string    `json:"ip" gorm:"type:varchar(128) not null"`
	Host       string    `json:"host" gorm:"type:varchar(45) not null"`
	Uri        string    `json:"uri" gorm:"type:varchar(255) not null"`
	Params     string    `json:"params" gorm:"type:text"`
	Cost       int64     `json:"cost" gorm:"type:bigint(20) not null;default:0"`
	Success    bool      `json:"success" gorm:"type:tinyint(1) not null;default:0"`
}
