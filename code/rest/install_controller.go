package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

//install apis. Only when installing period can be visited.
type InstallController struct {
	BaseController
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	matterDao         *MatterDao
	matterService     *MatterService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
	tableNames        []IBase
}

func (this *InstallController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if c, ok := b.(*MatterDao); ok {
		this.matterDao = c
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if c, ok := b.(*MatterService); ok {
		this.matterService = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}

	this.tableNames = []IBase{
		&Dashboard{},
		&Bridge{},
		&DownloadToken{},
		&Footprint{},
		&ImageCache{},
		&Matter{},
		&Preference{},
		&Session{},
		&Share{},
		&UploadToken{},
		&User{},
	}

}

func (this *InstallController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/install/verify"] = this.Wrap(this.Verify, USER_ROLE_GUEST)
	routeMap["/api/install/table/info/list"] = this.Wrap(this.TableInfoList, USER_ROLE_GUEST)
	routeMap["/api/install/create/table"] = this.Wrap(this.CreateTable, USER_ROLE_GUEST)
	routeMap["/api/install/admin/list"] = this.Wrap(this.AdminList, USER_ROLE_GUEST)
	routeMap["/api/install/create/admin"] = this.Wrap(this.CreateAdmin, USER_ROLE_GUEST)
	routeMap["/api/install/validate/admin"] = this.Wrap(this.ValidateAdmin, USER_ROLE_GUEST)
	routeMap["/api/install/finish"] = this.Wrap(this.Finish, USER_ROLE_GUEST)

	return routeMap
}

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

	mysqlUrl := util.GetMysqlUrl(mysqlPort, mysqlHost, mysqlSchema, mysqlUsername, mysqlPassword)

	this.logger.Info("Connect MySQL %s", mysqlUrl)

	var err error = nil
	db, err := gorm.Open("mysql", mysqlUrl)
	this.PanicError(err)

	db.LogMode(false)

	return db

}

func (this *InstallController) closeDbConnection(db *gorm.DB) {

	if db != nil {
		err := db.Close()
		if err != nil {
			this.logger.Error("occur error when close db. %v", err)
		}
	}
}

func (this *InstallController) getTableMeta(gormDb *gorm.DB, entity IBase) (bool, []*gorm.StructField, []*gorm.StructField) {

	db := gormDb.Unscoped()
	scope := db.NewScope(entity)

	tableName := scope.TableName()
	modelStruct := scope.GetModelStruct()
	allFields := modelStruct.StructFields
	var missingFields = make([]*gorm.StructField, 0)

	if !scope.Dialect().HasTable(tableName) {
		missingFields = append(missingFields, allFields...)

		return false, allFields, missingFields
	} else {

		for _, field := range allFields {
			if !scope.Dialect().HasColumn(tableName, field.DBName) {
				if field.IsNormal {
					missingFields = append(missingFields, field)
				}
			}
		}

		return true, allFields, missingFields
	}

}

func (this *InstallController) getTableMetaList(db *gorm.DB) []*InstallTableInfo {

	var installTableInfos []*InstallTableInfo

	for _, iBase := range this.tableNames {
		exist, allFields, missingFields := this.getTableMeta(db, iBase)
		installTableInfos = append(installTableInfos, &InstallTableInfo{
			Name:          iBase.TableName(),
			TableExist:    exist,
			AllFields:     allFields,
			MissingFields: missingFields,
		})
	}

	return installTableInfos
}

// validate table whether integrity. if not panic err.
func (this *InstallController) validateTableMetaList(tableInfoList []*InstallTableInfo) {

	for _, tableInfo := range tableInfoList {
		if tableInfo.TableExist {
			if len(tableInfo.MissingFields) != 0 {

				var strs []string
				for _, v := range tableInfo.MissingFields {
					strs = append(strs, v.DBName)
				}

				panic(result.BadRequest(fmt.Sprintf("table %s miss the following fields %v", tableInfo.Name, strs)))
			}
		} else {
			panic(result.BadRequest(tableInfo.Name + " table not exist"))
		}
	}

}

//Ping db.
func (this *InstallController) Verify(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	this.logger.Info("Ping DB")
	err := db.DB().Ping()
	this.PanicError(err)

	return this.Success("OK")
}

func (this *InstallController) TableInfoList(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	return this.Success(this.getTableMetaList(db))
}

func (this *InstallController) CreateTable(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	var installTableInfos []*InstallTableInfo

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	for _, iBase := range this.tableNames {

		//complete the missing fields or create table. use utf8 charset
		db1 := db.Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(iBase)
		this.PanicError(db1.Error)

		exist, allFields, missingFields := this.getTableMeta(db, iBase)
		installTableInfos = append(installTableInfos, &InstallTableInfo{
			Name:          iBase.TableName(),
			TableExist:    exist,
			AllFields:     allFields,
			MissingFields: missingFields,
		})

	}

	return this.Success(installTableInfos)

}

//get the list of admin.
func (this *InstallController) AdminList(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	var wp = &builder.WherePair{}

	wp = wp.And(&builder.WherePair{Query: "role = ?", Args: []interface{}{USER_ROLE_ADMINISTRATOR}})

	var users []*User
	db = db.Where(wp.Query, wp.Args...).Offset(0).Limit(10).Find(&users)

	this.PanicError(db.Error)

	return this.Success(users)
}

//create admin
func (this *InstallController) CreateAdmin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	adminUsername := request.FormValue("adminUsername")
	adminPassword := request.FormValue("adminPassword")

	//validate admin's username
	if m, _ := regexp.MatchString(USERNAME_PATTERN, adminUsername); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(adminPassword) < 6 {
		panic(result.BadRequest(`admin's password at least 6 chars'`))
	}

	//check whether duplicate
	var count2 int64
	db2 := db.Model(&User{}).Where("username = ?", adminUsername).Count(&count2)
	this.PanicError(db2.Error)
	if count2 > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, adminUsername))
	}

	user := &User{}
	timeUUID, _ := uuid.NewV4()
	user.Uuid = string(timeUUID.String())
	user.CreateTime = time.Now()
	user.UpdateTime = time.Now()
	user.LastTime = time.Now()
	user.Sort = time.Now().UnixNano() / 1e6
	user.Role = USER_ROLE_ADMINISTRATOR
	user.Username = adminUsername
	user.Password = util.GetBcrypt(adminPassword)
	user.SizeLimit = -1
	user.Status = USER_STATUS_OK

	db3 := db.Create(user)
	this.PanicError(db3.Error)

	return this.Success("OK")

}

//(if there is admin in db)Validate admin.
func (this *InstallController) ValidateAdmin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	adminUsername := request.FormValue("adminUsername")
	adminPassword := request.FormValue("adminPassword")

	if adminUsername == "" {
		panic(result.BadRequest(`admin's username cannot be null'`))
	}
	if len(adminPassword) < 6 {
		panic(result.BadRequest(`admin's password at least 6 chars'`))
	}

	var existUsernameUser = &User{}
	db = db.Where(&User{Username: adminUsername}).First(existUsernameUser)
	if db.Error != nil {
		panic(result.BadRequestI18n(request, i18n.UsernameNotExist, adminUsername))
	}

	if !util.MatchBcrypt(adminPassword, existUsernameUser.Password) {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError, adminUsername))
	}

	if existUsernameUser.Role != USER_ROLE_ADMINISTRATOR {
		panic(result.BadRequestI18n(request, i18n.UsernameIsNotAdmin, adminUsername))
	}

	return this.Success("OK")

}

//Finish the installation
func (this *InstallController) Finish(writer http.ResponseWriter, request *http.Request) *result.WebResult {

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

	//Recheck the db connection
	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	//Recheck the integrity of tables.
	tableMetaList := this.getTableMetaList(db)
	this.validateTableMetaList(tableMetaList)

	//At least one admin
	var count1 int64
	db1 := db.Model(&User{}).Where("role = ?", USER_ROLE_ADMINISTRATOR).Count(&count1)
	this.PanicError(db1.Error)
	if count1 == 0 {
		panic(result.BadRequest(`please config at least one admin user`))
	}

	//announce the config to write config to tank.json
	core.CONFIG.FinishInstall(mysqlPort, mysqlHost, mysqlSchema, mysqlUsername, mysqlPassword)

	//announce the context to broadcast the installation news to bean.
	core.CONTEXT.InstallOk()

	return this.Success("OK")
}
