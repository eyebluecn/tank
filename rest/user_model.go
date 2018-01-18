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
	Role      string    `json:"role"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Gender    string    `json:"gender"`
	City      string    `json:"city"`
	AvatarUrl string    `json:"avatarUrl"`
	LastIp    string    `json:"lastIp"`
	LastTime  time.Time `json:"lastTime"`
	SizeLimit int64     `json:"sizeLimit"`
	Status    string    `json:"status"`
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
