package rest

import (
	"time"
)

const (
	//游客身份
	USER_ROLE_GUEST = "GUEST"
	//普通注册用户
	USER_ROLE_USER = "USER"
	//管理员
	USER_ROLE_ADMINISTRATOR = "ADMINISTRATOR"
)

const (
	USER_GENDER_MALE    = "MALE"
	USER_GENDER_FEMALE  = "FEMALE"
	USER_GENDER_UNKNOWN = "UNKNOWN"
)

const (
	//正常状态
	USER_STATUS_OK       = "OK"
	//被禁用
	USER_STATUS_DISABLED = "DISABLED"
)

type User struct {
	Base
	Role      string    `json:"role" gorm:"type:varchar(45)"`
	Username  string    `json:"username" gorm:"type:varchar(45) not null;unique"`
	Password  string    `json:"-" gorm:"type:varchar(255)"`
	Email     string    `json:"email" gorm:"type:varchar(45) not null;unique"`
	Phone     string    `json:"phone" gorm:"type:varchar(45)"`
	Gender    string    `json:"gender" gorm:"type:varchar(45)"`
	City      string    `json:"city" gorm:"type:varchar(45)"`
	AvatarUrl string    `json:"avatarUrl" gorm:"type:varchar(255)"`
	LastIp    string    `json:"lastIp" gorm:"type:varchar(128)"`
	LastTime  time.Time `json:"lastTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	SizeLimit int64     `json:"sizeLimit" gorm:"type:bigint(20) not null;default:-1"`
	Status    string    `json:"status" gorm:"type:varchar(45)"`
}

// set User's table name to be `profiles`
func (User) TableName() string {
	return TABLE_PREFIX + "user"
}

//通过一个字符串获取性别
func GetGender(genderString string) string {
	if genderString == USER_GENDER_MALE || genderString == USER_GENDER_FEMALE || genderString == USER_GENDER_UNKNOWN {
		return genderString
	} else {
		return USER_GENDER_UNKNOWN
	}
}

//通过一个字符串获取角色
func GetRole(roleString string) string {
	if roleString == USER_ROLE_USER || roleString == USER_ROLE_ADMINISTRATOR {
		return roleString
	} else {
		return USER_ROLE_USER
	}
}
