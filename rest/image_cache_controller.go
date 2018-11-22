package rest

import (
	"net/http"
	"strconv"
	"strings"
)

type ImageCacheController struct {
	BaseController
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
}

//初始化方法
func (this *ImageCacheController) Init(context *Context) {
	this.BaseController.Init(context)

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := context.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = context.GetBean(this.imageCacheService)
	if b, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = b
	}

}

//注册自己的路由。
func (this *ImageCacheController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/image/cache/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/image/cache/delete/batch"] = this.Wrap(this.DeleteBatch, USER_ROLE_USER)
	routeMap["/api/image/cache/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/image/cache/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

//查看某个文件的详情。
func (this *ImageCacheController) Detail(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		return this.Error("文件的uuid必填")
	}

	imageCache := this.imageCacheService.Detail(uuid)

	//验证当前之人是否有权限查看这么详细。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		if imageCache.UserUuid != user.Uuid {
			panic("没有权限查看该文件")
		}
	}

	return this.Success(imageCache)

}

//按照分页的方式获取某个文件夹下文件和子文件夹的列表，通常情况下只有一页。
func (this *ImageCacheController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {

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

	pager := this.imageCacheDao.Page(page, pageSize, userUuid, sortArray)

	return this.Success(pager)
}

//删除一个文件
func (this *ImageCacheController) Delete(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		return this.Error("文件的uuid必填")
	}

	imageCache := this.imageCacheDao.FindByUuid(uuid)

	//判断文件的所属人是否正确
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && imageCache.UserUuid != user.Uuid {
		return this.Error(RESULT_CODE_UNAUTHORIZED)
	}

	this.imageCacheDao.Delete(imageCache)

	return this.Success("删除成功！")
}

//删除一系列文件。
func (this *ImageCacheController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		return this.Error("文件的uuids必填")
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		imageCache := this.imageCacheDao.FindByUuid(uuid)

		//判断文件的所属人是否正确
		user := this.checkUser(writer, request)
		if user.Role != USER_ROLE_ADMINISTRATOR && imageCache.UserUuid != user.Uuid {
			return this.Error(RESULT_CODE_UNAUTHORIZED)
		}

		this.imageCacheDao.Delete(imageCache)

	}

	return this.Success("删除成功！")
}
