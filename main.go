package main

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/support"
	_ "gorm.io/driver/mysql"
)

func main() {

	core.APPLICATION = &support.TankApplication{}
	core.APPLICATION.Start()

	//getlastmodified
	//Sat, 8 Jul 2023 12:08:15 GMT
	//dict := make(map[string]string)
	//dict["getlastmodified"] = "Sat, 8 Jul 2023 12:08:15 GMT"
	//marshal, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(dict)
	//if err != nil {
	//	return
	//}
	//println(marshal)
}
