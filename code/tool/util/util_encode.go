package util

import (
	"crypto/md5"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

//md5
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
