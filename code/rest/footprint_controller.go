package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"net/http"
)

type FootprintController struct {
	BaseController
	footprintDao     *FootprintDao
	footprintService *FootprintService
}

//初始化方法
func (this *FootprintController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	b = core.CONTEXT.GetBean(this.footprintService)
	if b, ok := b.(*FootprintService); ok {
		this.footprintService = b
	}

}

//注册自己的路由。
func (this *FootprintController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	return routeMap
}
