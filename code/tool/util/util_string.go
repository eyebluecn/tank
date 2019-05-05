package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

//human readable file size
func HumanFileSize(bytes int64) string {
	var thresh int64 = 1024

	if bytes < 0 {
		bytes = 0
	}
	if bytes < thresh {
		return fmt.Sprintf("%dB", bytes)
	}
	var units = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	var u = 0
	var tmp = float64(bytes)
	var standard = float64(thresh)
	for tmp >= standard && u < len(units)-1 {
		tmp /= float64(standard)
		u++
	}

	numStr := strconv.FormatFloat(tmp, 'f', 1, 64)

	return fmt.Sprintf("%s%s", numStr, units[u])
}

//get mysql url.
func GetMysqlUrl(
	mysqlPort int,
	mysqlHost string,
	mysqlSchema string,
	mysqlUsername string,
	mysqlPassword string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", mysqlUsername, mysqlPassword, mysqlHost, mysqlPort, mysqlSchema)
}

//get random number 4.
func RandomNumber4() string {
	return fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31()%10000)
}

//get random 4 string
func RandomString4() string {

	//0 and o, 1 and l are not easy to distinguish
	var letterRunes = []rune("abcdefghijkmnpqrstuvwxyz23456789")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]rune, 4)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}

	return string(b)
}
