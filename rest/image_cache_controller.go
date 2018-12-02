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
func (this *ImageCacheController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = CONTEXT.GetBean(this.imageCacheService)
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

//查看某个图片缓存的详情。
func (this *ImageCacheController) Detail(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		this.PanicBadRequest("图片缓存的uuid必填")
	}

	imageCache := this.imageCacheService.Detail(uuid)

	//验证当前之人是否有权限查看这么详细。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		if imageCache.UserUuid != user.Uuid {
			panic("没有权限查看该图片缓存")
		}
	}

	return this.Success(imageCache)

}

//按照分页的方式获取某个图片缓存夹下图片缓存和子图片缓存夹的列表，通常情况下只有一页。
func (this *ImageCacheController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {
	//如果是根目录，那么就传入root.
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")

	userUuid := request.FormValue("userUuid")
	matterUuid := request.FormValue("matterUuid")
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
			key:   "update_time",
			value: orderUpdateTime,
		},
		{
			key:   "sort",
			value: orderSort,
		},

		{
			key:   "size",
			value: orderSize,
		},
	}

	pager := this.imageCacheDao.Page(page, pageSize, userUuid, matterUuid, sortArray)

	return this.Success(pager)
}

//删除一个图片缓存
func (this *ImageCacheController) Delete(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		this.PanicBadRequest("图片缓存的uuid必填")
	}

	imageCache := this.imageCacheDao.FindByUuid(uuid)

	//判断图片缓存的所属人是否正确
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && imageCache.UserUuid != user.Uuid {
		this.PanicUnauthorized("没有权限")
	}

	this.imageCacheDao.Delete(imageCache)

	return this.Success("删除成功！")
}

//删除一系列图片缓存。
func (this *ImageCacheController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		this.PanicBadRequest("图片缓存的uuids必填")
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		imageCache := this.imageCacheDao.FindByUuid(uuid)

		//判断图片缓存的所属人是否正确
		user := this.checkUser(writer, request)
		if user.Role != USER_ROLE_ADMINISTRATOR && imageCache.UserUuid != user.Uuid {
			this.PanicUnauthorized("没有权限")
		}

		this.imageCacheDao.Delete(imageCache)

	}

	return this.Success("删除成功！")
}
