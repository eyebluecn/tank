package main

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/support"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("OSS Go SDK Version: ", oss.Version)
	core.APPLICATION = &support.TankApplication{}
	core.APPLICATION.Start()

}
