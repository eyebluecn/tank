package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

//把一个大小转变成方便读的格式
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

//获取MySQL的URL
func GetMysqlUrl(
	mysqlPort int,
	mysqlHost string,
	mysqlSchema string,
	mysqlUsername string,
	mysqlPassword string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", mysqlUsername, mysqlPassword, mysqlHost, mysqlPort, mysqlSchema)
}

//获取四位随机数字
func RandomNumber4() string {
	return fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31()%10000)
}

//获取四位随机数字
func RandomString4() string {

	//0和o，l和1难以区分，剔除掉
	var letterRunes = []rune("abcdefghijkmnpqrstuvwxyz23456789")

	b := make([]rune, 4)
	for i := range b {
		b[i] = letterRunes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(letterRunes))]
	}

	return string(b)
}
