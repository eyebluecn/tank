package main

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/support"
	_ "gorm.io/driver/mysql"
)

func main() {

	core.APPLICATION = &support.TankApplication{}
	core.APPLICATION.Start()

}
