package rest

import (
	"net/http"
)

type DashboardController struct {
	BaseController
	dashboardDao     *DashboardDao
	dashboardService *DashboardService
}

//初始化方法
func (this *DashboardController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.dashboardDao)
	if b, ok := b.(*DashboardDao); ok {
		this.dashboardDao = b
	}

	b = CONTEXT.GetBean(this.dashboardService)
	if b, ok := b.(*DashboardService); ok {
		this.dashboardService = b
	}

}

//注册自己的路由。
func (this *DashboardController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/dashboard/invoke/list"] = this.Wrap(this.InvokeList, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

//过去七天分时调用量
func (this *DashboardController) InvokeList(writer http.ResponseWriter, request *http.Request) *WebResult {

	return this.Success("")

}
