package main

import (
	"fmt"
	"github.com/eyebluecn/tank/code/config"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/rest"
	"github.com/eyebluecn/tank/code/support"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

func main() {

	//日志第一优先级保障
	tankLogger := &support.TankLogger{}
	tankLogger.Init()
	defer tankLogger.Destroy()
	core.LOGGER = tankLogger

	//装载配置文件，这个决定了是否需要执行安装过程
	config.CONFIG.Init()

	//全局运行的上下文
	tankContext := &rest.Context{}
	tankContext.Init()
	defer tankContext.Destroy()
	core.CONTEXT = tankContext

	http.Handle("/", core.CONTEXT)

	core.LOGGER.Info("App started at http://localhost:%v", config.CONFIG.ServerPort)

	dotPort := fmt.Sprintf(":%v", config.CONFIG.ServerPort)
	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
