package main

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/support"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	core.APPLICATION = &support.TankApplication{}
	core.APPLICATION.Start()

}
