package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

/**
 *
 * WebDav协议文档
 * https://tools.ietf.org/html/rfc4918
 * http://www.webdav.org/specs/rfc4918.html
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

//通过BasicAuth的方式授权。
func (this *DavController) CheckCurrentUser(writer http.ResponseWriter, request *http.Request) *User {

	username, password, ok := request.BasicAuth()
	if !ok {
		//要求前端使用Basic的形式授权
		writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		panic(ConstWebResult(CODE_WRAPPER_LOGIN))

	}

	user := this.userDao.FindByUsername(username)
	if user == nil {
		this.PanicBadRequest("邮箱或密码错误")
	} else {
		if !MatchBcrypt(password, user.Password) {
			this.PanicBadRequest("邮箱或密码错误")
		}
	}

	return user
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
	pattern := fmt.Sprintf(`^%s(.*)$`, WEBDAV_PREFFIX)
	reg := regexp.MustCompile(pattern)
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

	/*打印所有HEADER以及请求参数*/

	fmt.Printf("\n------ 请求： %s  --  %s  ------\n", request.URL, subPath)

	fmt.Printf("\n------Method：------\n")
	fmt.Println(request.Method)

	fmt.Printf("\n------Header：------\n")
	for key, value := range request.Header {
		fmt.Printf("%s = %s\n", key, value)
	}

	fmt.Printf("\n------请求参数：------\n")
	for key, value := range request.Form {
		fmt.Printf("%s = %s\n", key, value)
	}

	fmt.Printf("\n------Body：------\n")
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println("读取body时出错" + err.Error())
	}
	fmt.Println(string(body))

	fmt.Println("------------------")

	//获取请求者
	user := this.CheckCurrentUser(writer, request)

	method := request.Method
	if method == "PROPFIND" {
		//列出文件夹或者目录详情
		this.davService.HandlePropfind(writer, request, user, subPath)

	} else if method == "GET" {
		//请求文件详情（下载）
		this.davService.HandleGet(writer, request, user, subPath)

	} else if method == "DELETE" {
		//删除文件
		this.davService.HandleDelete(writer, request, user, subPath)

	} else {

		this.PanicBadRequest("该方法还不支持。%s", method)
	}

}
