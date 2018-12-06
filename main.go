package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"tank/rest"
)

func main() {

	//日志第一优先级保障
	rest.LOGGER.Init()
	defer rest.LOGGER.Destroy()

	//装载配置文件，这个决定了是否需要执行安装过程
	rest.CONFIG.Init()

	//全局运行的上下文
	rest.CONTEXT.Init()
	defer rest.CONTEXT.Destroy()

	http.Handle("/", rest.CONTEXT.Router)

	rest.LOGGER.Info("App started at http://localhost:%v", rest.CONFIG.ServerPort)

	dotPort := fmt.Sprintf(":%v", rest.CONFIG.ServerPort)
	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
