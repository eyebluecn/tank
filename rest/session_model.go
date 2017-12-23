package rest

import (
	"time"
)

type Session struct {
	Base
	Authentication string    `json:"authentication"`
	UserUuid       string    `json:"userUuid"`
	Ip             string    `json:"ip"`
	ExpireTime     time.Time `json:"expireTime"`
}

// set User's table name to be `profiles`
func (Session) TableName() string {
	return TABLE_PREFIX + "session"
}
