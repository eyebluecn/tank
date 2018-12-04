package rest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"go/build"
	"io/ioutil"
	"net/http"
	"strconv"
)

//安装程序的接口，只有安装阶段可以访问。
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
	routeMap["/api/install/table/info/list"] = this.Wrap(this.InstallTableInfoList, USER_ROLE_GUEST)

	return routeMap
}

//获取数据库连接
func (this *InstallController) openDbConnection(writer http.ResponseWriter, request *http.Request) *gorm.DB {
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

	this.logger.Info("连接MySQL %s", mysqlUrl)

	var err error = nil
	db, err := gorm.Open("mysql", mysqlUrl)
	this.PanicError(err)

	return db

}

//关闭数据库连接
func (this *InstallController) closeDbConnection(db *gorm.DB) {

	if db != nil {
		err := db.Close()
		if err != nil {
			this.logger.Error("关闭数据库连接出错 %v", err)
		}
	}
}

//根据表名获取建表SQL语句
func (this *InstallController) getCreateSQLFromFile(tableName string) string {

	//1. 从当前安装目录db下去寻找建表文件。
	homePath := GetHomePath()
	filePath := homePath + "/db/" + tableName + ".sql"
	exists, err := PathExists(filePath)
	if err != nil {
		this.PanicServer("从安装目录判断建表语句文件是否存在时出错！")
	}

	//2. 从GOPATH下面去找，因为可能是开发环境
	if !exists {

		this.logger.Info("GOPATH = %s", build.Default.GOPATH)

		filePath1 := filePath
		filePath = build.Default.GOPATH + "/src/tank/build/db/" + tableName + ".sql"
		exists, err = PathExists(filePath)
		if err != nil {
			this.PanicServer("从GOPATH判断建表语句文件是否存在时出错！")
		}

		if !exists {
			this.PanicServer(fmt.Sprintf("%s 或 %s 均不存在，请检查你的安装情况。", filePath1, filePath))
		}
	}

	//读取文件内容.
	bytes, err := ioutil.ReadFile(filePath)
	this.PanicError(err)

	return string(bytes)
}

//根据表名获取建表SQL语句
func (this *InstallController) getCreateSQLFromDb(db *gorm.DB, base IBase) (bool, string) {

	var hasTable = true
	var tableName = base.TableName()
	hasTable = db.HasTable(base)
	if !hasTable {
		return false, ""
	}

	// Scan
	type Result struct {
		Table       string
		CreateTable string
	}

	//读取建表语句。
	var result = &Result{}
	db1 := db.Exec("SHOW CREATE TABLE " + tableName).Scan(result)
	this.PanicError(db1.Error)

	return true, result.CreateTable
}

//验证数据库连接
func (this *InstallController) Verify(writer http.ResponseWriter, request *http.Request) *WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	this.logger.Info("Ping一下数据库")
	err := db.DB().Ping()
	this.PanicError(err)

	return this.Success("OK")
}

//获取需要安装的数据库表
func (this *InstallController) InstallTableInfoList(writer http.ResponseWriter, request *http.Request) *WebResult {

	var tableNames = []IBase{&Dashboard{}, &DownloadToken{}, &Footprint{}, &ImageCache{}, &Matter{}, &Preference{}, &Session{}, UploadToken{}, &User{}}
	var installTableInfos []*InstallTableInfo

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	for _, iBase := range tableNames {

		exist, sql := this.getCreateSQLFromDb(db, iBase)
		installTableInfos = append(installTableInfos, &InstallTableInfo{
			Name:           iBase.TableName(),
			CreateSql:      this.getCreateSQLFromFile(iBase.TableName()),
			TableExist:     exist,
			ExistCreateSql: sql,
		})

	}

	return this.Success(installTableInfos)

}
