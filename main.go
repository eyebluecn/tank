package main

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/support"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

func main() {

	//第一步。命令行工具处理
	tankCommand := &support.TankCommand{}
	core.COMMAND = tankCommand
	if core.COMMAND.Cli() {
		return
	}

	//第二步。日志
	tankLogger := &support.TankLogger{}
	core.LOGGER = tankLogger
	tankLogger.Init()
	defer tankLogger.Destroy()

	//第三步。配置
	tankConfig := &support.TankConfig{}
	core.CONFIG = tankConfig
	tankConfig.Init()

	//第四步。全局运行的上下文
	tankContext := &support.TankContext{}
	core.CONTEXT = tankContext
	tankContext.Init()
	defer tankContext.Destroy()

	//第五步。启动http服务
	http.Handle("/", core.CONTEXT)
	core.LOGGER.Info("App started at http://localhost:%v", core.CONFIG.ServerPort())

	dotPort := fmt.Sprintf(":%v", core.CONFIG.ServerPort())
	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
