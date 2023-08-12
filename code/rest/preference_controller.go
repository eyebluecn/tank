package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"strconv"
)

type PreferenceController struct {
	BaseController
	preferenceDao     *PreferenceDao
	matterDao         *MatterDao
	matterService     *MatterService
	preferenceService *PreferenceService
	taskService       *TaskService
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
	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}
	b = core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

	b = core.CONTEXT.GetBean(this.taskService)
	if b, ok := b.(*TaskService); ok {
		this.taskService = b
	}

}

func (this *PreferenceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/preference/ping"] = this.Wrap(this.Ping, USER_ROLE_GUEST)
	routeMap["/api/preference/fetch"] = this.Wrap(this.Fetch, USER_ROLE_GUEST)
	routeMap["/api/preference/edit"] = this.Wrap(this.Edit, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/edit/preview/config"] = this.Wrap(this.EditPreviewConfig, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/edit/scan/config"] = this.Wrap(this.EditScanConfig, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/scan/once"] = this.Wrap(this.ScanOnce, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/preference/system/cleanup"] = this.Wrap(this.SystemCleanup, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

// ping the application. Return current version.
func (this *PreferenceController) Ping(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	return this.Success(core.VERSION)

}

func (this *PreferenceController) Fetch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	preference := this.preferenceService.Fetch()

	return this.Success(preference)
}

// edit basic info.
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
	deletedKeepDaysStr := request.FormValue("deletedKeepDays")

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

	var deletedKeepDays int64 = 0
	if deletedKeepDaysStr == "" {
		panic(result.BadRequest("deletedKeepDays cannot be null"))
	} else {
		intDeletedKeepDays, err := strconv.Atoi(deletedKeepDaysStr)
		this.PanicError(err)
		deletedKeepDays = int64(intDeletedKeepDays)

		if deletedKeepDays < 0 {
			panic(result.BadRequest("deletedKeepDays cannot less than 0"))
		}
	}

	var allowRegister = false
	if allowRegisterStr == TRUE {
		allowRegister = true
	}

	preference := this.preferenceDao.Fetch()
	oldDeletedKeepDays := preference.DeletedKeepDays
	preference.Name = name
	preference.LogoUrl = logoUrl
	preference.FaviconUrl = faviconUrl
	preference.Copyright = copyright
	preference.Record = record
	preference.DownloadDirMaxSize = downloadDirMaxSize
	preference.DownloadDirMaxNum = downloadDirMaxNum
	preference.DefaultTotalSizeLimit = defaultTotalSizeLimit
	preference.AllowRegister = allowRegister
	preference.DeletedKeepDays = deletedKeepDays

	preference = this.preferenceService.Save(preference)

	//if changed the bin strategy. then trigger once.
	if oldDeletedKeepDays != deletedKeepDays {
		this.matterService.CleanExpiredDeletedMatters()
	}

	return this.Success(preference)
}

// edit preview config.
func (this *PreferenceController) EditPreviewConfig(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	previewConfig := request.FormValue("previewConfig")

	preference := this.preferenceDao.Fetch()
	preference.PreviewConfig = previewConfig

	preference = this.preferenceService.Save(preference)

	return this.Success(preference)
}

func (this *PreferenceController) EditScanConfig(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	scanConfigStr := request.FormValue("scanConfig")
	if scanConfigStr == "" {
		panic(result.BadRequest("scanConfig cannot be null"))
	}

	preference := this.preferenceDao.Fetch()

	scanConfig := &ScanConfig{}
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(scanConfigStr), &scanConfig)
	if err != nil {
		panic(err)
	}

	//validate the scan config.
	if scanConfig.Enable {
		//validate cron.
		if !util.ValidateCron(scanConfig.Cron) {
			panic(result.CustomWebResultI18n(request, result.SHARE_CODE_ERROR, i18n.CronValidateError))
		}

		//validate scope.
		if scanConfig.Scope == SCAN_SCOPE_CUSTOM {
			if len(scanConfig.SpaceNames) == 0 {
				panic(result.BadRequest("scope cannot be null"))
			}
		} else if scanConfig.Scope == SCAN_SCOPE_ALL {

		} else {
			panic(result.BadRequest("cannot recognize scope %s", scanConfig.Scope))
		}
	}

	preference.ScanConfig = scanConfigStr
	preference = this.preferenceService.Save(preference)

	//reinit the scan task.
	this.taskService.InitScanTask()

	return this.Success(preference)
}

// scan immediately according the current config.
func (this *PreferenceController) ScanOnce(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	this.taskService.doScanTask()
	return this.Success("OK")
}

// cleanup system data.
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
