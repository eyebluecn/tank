package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"tank/rest"
	"tank/rest/config"
	"tank/rest/logger"
)

func main() {

	//日志第一优先级保障
	logger.LOGGER.Init()
	defer logger.LOGGER.Destroy()

	//装载配置文件，这个决定了是否需要执行安装过程
	config.CONFIG.Init()

	//全局运行的上下文
	rest.CONTEXT.Init()
	defer rest.CONTEXT.Destroy()

	http.Handle("/", rest.CONTEXT.Router)

	logger.LOGGER.Info("App started at http://localhost:%v", config.CONFIG.ServerPort)

	dotPort := fmt.Sprintf(":%v", config.CONFIG.ServerPort)
	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
