package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
)

type DashboardController struct {
	BaseController
	dashboardDao     *DashboardDao
	dashboardService *DashboardService
}

func (this *DashboardController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.dashboardDao)
	if b, ok := b.(*DashboardDao); ok {
		this.dashboardDao = b
	}

	b = core.CONTEXT.GetBean(this.dashboardService)
	if b, ok := b.(*DashboardService); ok {
		this.dashboardService = b
	}
}

func (this *DashboardController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/dashboard/page"] = this.Wrap(this.Page, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/dashboard/active/ip/top10"] = this.Wrap(this.ActiveIpTop10, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/dashboard/etl"] = this.Wrap(this.Etl, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

func (this *DashboardController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

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

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
		{
			Key:   "update_time",
			Value: orderUpdateTime,
		},
		{
			Key:   "sort",
			Value: orderSort,
		},
		{
			Key:   "dt",
			Value: orderDt,
		},
	}

	pager := this.dashboardDao.Page(page, pageSize, "", sortArray)

	return this.Success(pager)
}

func (this *DashboardController) ActiveIpTop10(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	//TODO:
	list := this.dashboardDao.ActiveIpTop10()
	return this.Success(list)
}

func (this *DashboardController) Etl(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	this.dashboardService.Etl()
	return this.Success("OK")
}
