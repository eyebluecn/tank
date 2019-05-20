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
	matterDao         *MatterDao
	preferenceService *PreferenceService
}

func (this *PreferenceController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*PreferenceDao); ok {
		this.preferenceDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}
	b = core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

}

func (this *PreferenceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/preference/ping"] = this.Wrap(this.Ping, USER_ROLE_GUEST)
	routeMap["/api/preference/fetch"] = this.Wrap(this.Fetch, USER_ROLE_GUEST)
	routeMap["/api/preference/edit"] = this.Wrap(this.Edit, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/system/cleanup"] = this.Wrap(this.SystemCleanup, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/migrate20to30"] = this.Wrap(this.Migrate20to30, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

//ping the application. Return current version.
func (this *PreferenceController) Ping(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	return this.Success(core.VERSION)

}

func (this *PreferenceController) Fetch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	preference := this.preferenceService.Fetch()

	return this.Success(preference)
}

func (this *PreferenceController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	name := request.FormValue("name")

	logoUrl := request.FormValue("logoUrl")
	faviconUrl := request.FormValue("faviconUrl")
	copyright := request.FormValue("copyright")
	record := request.FormValue("record")
	downloadDirMaxSizeStr := request.FormValue("downloadDirMaxSize")
	downloadDirMaxNumStr := request.FormValue("downloadDirMaxNum")
	defaultTotalSizeLimitStr := request.FormValue("defaultTotalSizeLimit")
	allowRegisterStr := request.FormValue("allowRegister")

	if name == "" {
		panic(result.BadRequest("name cannot be null"))
	}

	var downloadDirMaxSize int64 = 0
	if downloadDirMaxSizeStr == "" {
		panic(result.BadRequest("downloadDirMaxSize cannot be null"))
	} else {
		intDownloadDirMaxSize, err := strconv.Atoi(downloadDirMaxSizeStr)
		this.PanicError(err)
		downloadDirMaxSize = int64(intDownloadDirMaxSize)
	}

	var downloadDirMaxNum int64 = 0
	if downloadDirMaxNumStr == "" {
		panic(result.BadRequest("downloadDirMaxNum cannot be null"))
	} else {
		intDownloadDirMaxNum, err := strconv.Atoi(downloadDirMaxNumStr)
		this.PanicError(err)
		downloadDirMaxNum = int64(intDownloadDirMaxNum)
	}

	var defaultTotalSizeLimit int64 = 0
	if defaultTotalSizeLimitStr == "" {
		panic(result.BadRequest("defaultTotalSizeLimit cannot be null"))
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

	//reset the preference cache
	this.preferenceService.Reset()

	return this.Success(preference)
}

//cleanup system data.
func (this *PreferenceController) SystemCleanup(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	user := this.checkUser(request)
	password := request.FormValue("password")

	if !util.MatchBcrypt(password, user.Password) {
		panic(result.BadRequest("password error"))
	}

	//this will trigger every bean to cleanup.
	core.CONTEXT.Cleanup()

	return this.Success("OK")
}

//migrate 2.0's db data and file data to 3.0
func (this *PreferenceController) Migrate20to30(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	this.logger.Info("start migrating from 2.0 to 3.0")

	this.preferenceService.Migrate20to30(writer, request)
	return this.Success("OK")
}
