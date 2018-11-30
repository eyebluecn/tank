package rest

import (
	"net/http"
	"strconv"
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
	b := CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	b = CONTEXT.GetBean(this.footprintService)
	if b, ok := b.(*FootprintService); ok {
		this.footprintService = b
	}

}

//注册自己的路由。
func (this *FootprintController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/footprint/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/footprint/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/footprint/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

//查看详情。
func (this *FootprintController) Detail(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		this.PanicBadRequest("图片缓存的uuid必填")
	}

	footprint := this.footprintService.Detail(uuid)

	//验证当前之人是否有权限查看这么详细。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		if footprint.UserUuid != user.Uuid {
			panic("没有权限查看该图片缓存")
		}
	}

	return this.Success(footprint)

}

//按照分页的方式查询
func (this *FootprintController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {

	//如果是根目录，那么就传入root.
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	userUuid := request.FormValue("userUuid")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderSize := request.FormValue("orderSize")

	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}

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
			key:   "size",
			value: orderSize,
		},
	}

	pager := this.footprintDao.Page(page, pageSize, userUuid, sortArray)

	return this.Success(pager)
}

//删除一条记录
func (this *FootprintController) Delete(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		this.PanicBadRequest("uuid必填")
	}

	footprint := this.footprintDao.FindByUuid(uuid)

	if footprint != nil {
		this.footprintDao.Delete(footprint)
	}

	return this.Success("删除成功！")
}
