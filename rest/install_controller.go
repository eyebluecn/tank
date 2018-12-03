package rest

import (
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
)

type InstallController struct {
	BaseController
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	matterDao         *MatterDao
	matterService     *MatterService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
}

//初始化方法
func (this *InstallController) Init() {
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

}

//注册自己的路由。
func (this *InstallController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/install/verify"] = this.Wrap(this.Verify, USER_ROLE_GUEST)
	routeMap["/api/install/table/dashboard"] = this.Wrap(this.InstallTableDashboard, USER_ROLE_GUEST)

	return routeMap
}

//验证数据库连接
func (this *InstallController) Verify(writer http.ResponseWriter, request *http.Request) *WebResult {

	mysqlPortStr := request.FormValue("mysqlPort")
	mysqlHost := request.FormValue("mysqlHost")
	mysqlSchema := request.FormValue("mysqlSchema")
	mysqlUsername := request.FormValue("mysqlUsername")
	mysqlPassword := request.FormValue("mysqlPassword")

	var mysqlPort int
	if mysqlPortStr != "" {
		tmp, err := strconv.Atoi(mysqlPortStr)
		this.PanicError(err)
		mysqlPort = tmp
	}

	mysqlUrl := GetMysqlUrl(mysqlPort, mysqlHost, mysqlSchema, mysqlUsername, mysqlPassword)
	this.logger.Info("验证MySQL连接性 %s", mysqlUrl)

	var err error = nil
	db, err := gorm.Open("mysql", mysqlUrl)
	this.PanicError(err)

	this.logger.Info("Ping一下数据库")
	err = db.DB().Ping()
	this.PanicError(err)

	return this.Success("OK")

}

//安装dashboard表
func (this *InstallController) InstallTableDashboard(writer http.ResponseWriter, request *http.Request) *WebResult {

	return this.Success("")

}
