package rest

import (
	"net/http"
	"strconv"
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
	routeMap["/api/dashboard/page"] = this.Wrap(this.Page, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/dashboard/active/ip/top10"] = this.Wrap(this.ActiveIpTop10, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

//过去七天分时调用量
func (this *DashboardController) InvokeList(writer http.ResponseWriter, request *http.Request) *WebResult {

	return this.Success("")

}

//按照分页的方式获取某个图片缓存夹下图片缓存和子图片缓存夹的列表，通常情况下只有一页。
func (this *DashboardController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {

	//如果是根目录，那么就传入root.
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")
	orderDt := request.FormValue("orderDt")

	var page int
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	pageSize := 200
	if pageSizeStr != "" {
		tmp, err := strconv.Atoi(pageSizeStr)
		if err == nil {
			pageSize = tmp
		}
	}

	sortArray := []OrderPair{
		{
			key:   "create_time",
			value: orderCreateTime,
		},
		{
			key:   "update_time",
			value: orderUpdateTime,
		},
		{
			key:   "sort",
			value: orderSort,
		},
		{
			key:   "dt",
			value: orderDt,
		},
	}

	pager := this.dashboardDao.Page(page, pageSize, "", sortArray)

	return this.Success(pager)
}


func (this *DashboardController) ActiveIpTop10(writer http.ResponseWriter, request *http.Request) *WebResult {
	list := this.dashboardDao.ActiveIpTop10()
	return this.Success(list)
}
