package rest

import (
	"net/http"
	"tank/rest/result"
	"tank/rest/tool"
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
	b := CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*PreferenceDao); ok {
		this.preferenceDao = b
	}

	b = CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

}

//注册自己的路由。
func (this *PreferenceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/preference/fetch"] = this.Wrap(this.Fetch, USER_ROLE_GUEST)
	routeMap["/api/preference/edit"] = this.Wrap(this.Edit, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/system_cleanup"] = this.Wrap(this.SystemCleanup, USER_ROLE_ADMINISTRATOR)

	return routeMap
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
	footerLine1 := request.FormValue("footerLine1")
	footerLine2 := request.FormValue("footerLine2")
	showAlienStr := request.FormValue("showAlien")

	preference := this.preferenceDao.Fetch()
	preference.Name = name
	preference.LogoUrl = logoUrl
	preference.FaviconUrl = faviconUrl
	preference.FooterLine1 = footerLine1
	preference.FooterLine2 = footerLine2
	if showAlienStr == TRUE {
		preference.ShowAlien = true
	} else if showAlienStr == FALSE {
		preference.ShowAlien = false
	}

	preference = this.preferenceDao.Save(preference)

	//重置缓存中的偏好
	this.preferenceService.Reset()

	return this.Success(preference)
}

//清扫系统，所有数据全部丢失。一定要非常慎点，非常慎点！只在系统初始化的时候点击！
func (this *PreferenceController) SystemCleanup(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	user := this.checkUser(writer, request)
	password := request.FormValue("password")

	if !tool.MatchBcrypt(password, user.Password) {
		this.PanicBadRequest("密码错误，不能执行！")
	}

	for _, bean := range CONTEXT.BeanMap {
		bean.Cleanup()
	}

	return this.Success("OK")
}
