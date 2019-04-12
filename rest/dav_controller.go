package rest

import (
	"net/http"
	"regexp"
	"tank/rest/dav"
)

/**
 *
 * WebDav协议文档
 * https://tools.ietf.org/html/rfc4918
 *
 */
type DavController struct {
	BaseController
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	matterDao         *MatterDao
	matterService     *MatterService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
	davService        *DavService
}

//初始化方法
func (this *DavController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean.
	b := CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = CONTEXT.GetBean(this.matterDao)
	if c, ok := b.(*MatterDao); ok {
		this.matterDao = c
	}

	b = CONTEXT.GetBean(this.matterService)
	if c, ok := b.(*MatterService); ok {
		this.matterService = c
	}

	b = CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}

	b = CONTEXT.GetBean(this.davService)
	if c, ok := b.(*DavService); ok {
		this.davService = c
	}
}

//注册自己的路由。
func (this *DavController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	return routeMap
}

//处理一些特殊的接口，比如参数包含在路径中,一般情况下，controller不将参数放在url路径中
func (this *DavController) HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool) {

	path := request.URL.Path

	//匹配 /api/dav{subPath}
	reg := regexp.MustCompile(`^/api/dav(.*)$`)
	strs := reg.FindStringSubmatch(path)
	if len(strs) == 2 {
		var f = func(writer http.ResponseWriter, request *http.Request) {
			this.Index(writer, request, strs[1])
		}
		return f, true
	}

	return nil, false
}


//完成系统安装
func (this *DavController) Index(writer http.ResponseWriter, request *http.Request, subPath string) {

	this.logger.Info("请求访问来了：%s %s", request.RequestURI, subPath)

	if request.Method == "PROPFIND" {

		this.davService.HandlePropfind(writer, request)

	} else {

		handler := &dav.Handler{
			FileSystem: dav.Dir("/Users/fusu/d/group/golang/src/tank/tmp/dav"),
			LockSystem: dav.NewMemLS(),
		}

		handler.ServeHTTP(writer, request)

	}

}
