package rest

import (
	"github.com/eyebluecn/tank/code/config"
	"time"
)

type Session struct {
	Base
	UserUuid   string    `json:"userUuid" gorm:"type:char(36)"`
	Ip         string    `json:"ip" gorm:"type:varchar(128) not null"`
	ExpireTime time.Time `json:"expireTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
}

// set User's table name to be `profiles`
func (this *Session) TableName() string {
	return config.TABLE_PREFIX + "session"
}
