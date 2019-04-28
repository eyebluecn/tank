package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
)

type BridgeController struct {
	BaseController
	bridgeDao     *BridgeDao
	shareDao      *ShareDao
	bridgeService *BridgeService
}

//初始化方法
func (this *BridgeController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}
	b = core.CONTEXT.GetBean(this.bridgeService)
	if b, ok := b.(*BridgeService); ok {
		this.bridgeService = b
	}

}

//注册自己的路由。
func (this *BridgeController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/bridge/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

//按照分页的方式查询
func (this *BridgeController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//如果是根目录，那么就传入root.
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	shareUuid := request.FormValue("shareUuid")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderSize := request.FormValue("orderSize")

	share := this.shareDao.CheckByUuid(shareUuid)

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
			Key:   "size",
			Value: orderSize,
		},
	}

	pager := this.bridgeDao.Page(page, pageSize, share.Uuid, sortArray)

	return this.Success(pager)
}
