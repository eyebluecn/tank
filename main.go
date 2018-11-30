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

	//将运行时参数装填到config中去。
	rest.PrepareConfigs()

	//全局运行的上下文
	rest.CONTEXT.Init()
	defer rest.CONTEXT.Destroy()
	
	http.Handle("/", rest.CONTEXT.Router)

	dotPort := fmt.Sprintf(":%v", rest.CONFIG.ServerPort)

	rest.LOGGER.Info("App started at http://localhost:%v", rest.CONFIG.ServerPort)

	err1 := http.ListenAndServe(dotPort, nil)
	if err1 != nil {
		log.Fatal("ListenAndServe: ", err1)
	}
}
