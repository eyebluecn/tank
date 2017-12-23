package rest

import (
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"crypto/md5"
)

//给密码字符串加密
func GetMd5(raw string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(raw)))
}

func GetBcrypt(raw string) string {

	password := []byte(raw)
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword)
}

func MatchBcrypt(raw string, bcryptStr string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(bcryptStr), []byte(raw))

	return err == nil

}
