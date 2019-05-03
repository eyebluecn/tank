package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"strconv"
)

type PreferenceController struct {
	BaseController
	preferenceDao     *PreferenceDao
	preferenceService *PreferenceService
}

//初始化方法
func (this *PreferenceController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*PreferenceDao); ok {
		this.preferenceDao = b
	}

	b = core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

}

//注册自己的路由。
func (this *PreferenceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/preference/ping"] = this.Wrap(this.Ping, USER_ROLE_GUEST)
	routeMap["/api/preference/fetch"] = this.Wrap(this.Fetch, USER_ROLE_GUEST)
	routeMap["/api/preference/edit"] = this.Wrap(this.Edit, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/system/cleanup"] = this.Wrap(this.SystemCleanup, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

//简单验证蓝眼云盘服务是否已经启动了。
func (this *PreferenceController) Ping(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	return this.Success(nil)

}

//查看某个偏好设置的详情。
func (this *PreferenceController) Fetch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	preference := this.preferenceService.Fetch()

	return this.Success(preference)
}

//修改
func (this *PreferenceController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//验证参数。
	name := request.FormValue("name")
	if name == "" {
		panic("name参数必填")
	}

	logoUrl := request.FormValue("logoUrl")
	faviconUrl := request.FormValue("faviconUrl")
	copyright := request.FormValue("copyright")
	record := request.FormValue("record")
	downloadDirMaxSizeStr := request.FormValue("downloadDirMaxSize")
	downloadDirMaxNumStr := request.FormValue("downloadDirMaxNum")
	defaultTotalSizeLimitStr := request.FormValue("defaultTotalSizeLimit")
	allowRegisterStr := request.FormValue("allowRegister")

	var downloadDirMaxSize int64 = 0
	if downloadDirMaxSizeStr == "" {
		panic("文件下载大小限制必填！")
	} else {
		intDownloadDirMaxSize, err := strconv.Atoi(downloadDirMaxSizeStr)
		this.PanicError(err)
		downloadDirMaxSize = int64(intDownloadDirMaxSize)
	}

	var downloadDirMaxNum int64 = 0
	if downloadDirMaxNumStr == "" {
		panic("文件下载数量限制必填！")
	} else {
		intDownloadDirMaxNum, err := strconv.Atoi(downloadDirMaxNumStr)
		this.PanicError(err)
		downloadDirMaxNum = int64(intDownloadDirMaxNum)
	}

	var defaultTotalSizeLimit int64 = 0
	if defaultTotalSizeLimitStr == "" {
		panic("用户默认总限制！")
	} else {
		intDefaultTotalSizeLimit, err := strconv.Atoi(defaultTotalSizeLimitStr)
		this.PanicError(err)
		defaultTotalSizeLimit = int64(intDefaultTotalSizeLimit)
	}

	var allowRegister = false
	if allowRegisterStr == TRUE {
		allowRegister = true
	}

	preference := this.preferenceDao.Fetch()
	preference.Name = name
	preference.LogoUrl = logoUrl
	preference.FaviconUrl = faviconUrl
	preference.Copyright = copyright
	preference.Record = record
	preference.DownloadDirMaxSize = downloadDirMaxSize
	preference.DownloadDirMaxNum = downloadDirMaxNum
	preference.DefaultTotalSizeLimit = defaultTotalSizeLimit
	preference.AllowRegister = allowRegister

	preference = this.preferenceDao.Save(preference)

	//重置缓存中的偏好
	this.preferenceService.Reset()

	return this.Success(preference)
}

//清扫系统，所有数据全部丢失。一定要非常慎点，非常慎点！只在系统初始化的时候点击！
func (this *PreferenceController) SystemCleanup(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	user := this.checkUser(request)
	password := request.FormValue("password")

	if !util.MatchBcrypt(password, user.Password) {
		panic(result.BadRequest("密码错误，不能执行！"))
	}

	//清空系统
	core.CONTEXT.Cleanup()

	return this.Success("OK")
}
