package core

import "net/http"

type Controller interface {
	Bean
	//注册自己固定的路由。
	RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request)
	//处理一些特殊的路由。
	HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool)
}
