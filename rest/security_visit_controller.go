package rest

import (
	"net/http"
	"strconv"
)

type SecurityVisitController struct {
	BaseController
	securityVisitDao     *SecurityVisitDao
	securityVisitService *SecurityVisitService
}

//初始化方法
func (this *SecurityVisitController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.securityVisitDao)
	if b, ok := b.(*SecurityVisitDao); ok {
		this.securityVisitDao = b
	}

	b = CONTEXT.GetBean(this.securityVisitService)
	if b, ok := b.(*SecurityVisitService); ok {
		this.securityVisitService = b
	}

}

//注册自己的路由。
func (this *SecurityVisitController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/security/visit/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/security/visit/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/security/visit/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

//查看详情。
func (this *SecurityVisitController) Detail(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		return this.Error("图片缓存的uuid必填")
	}

	securityVisit := this.securityVisitService.Detail(uuid)

	//验证当前之人是否有权限查看这么详细。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		if securityVisit.UserUuid != user.Uuid {
			panic("没有权限查看该图片缓存")
		}
	}

	return this.Success(securityVisit)

}

//按照分页的方式查询
func (this *SecurityVisitController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {

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

	pager := this.securityVisitDao.Page(page, pageSize, userUuid, sortArray)

	return this.Success(pager)
}

//删除一条记录
func (this *SecurityVisitController) Delete(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		return this.Error("图片缓存的uuid必填")
	}

	securityVisit := this.securityVisitDao.FindByUuid(uuid)

	//判断图片缓存的所属人是否正确
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && securityVisit.UserUuid != user.Uuid {
		return this.Error(CODE_WRAPPER_UNAUTHORIZED)
	}

	this.securityVisitDao.Delete(securityVisit)

	return this.Success("删除成功！")
}
