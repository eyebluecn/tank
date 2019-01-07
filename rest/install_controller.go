package rest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/json-iterator/go"
	"github.com/nu7hatch/gouuid"
	"go/build"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
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
	routeMap["/api/install/table/info/list"] = this.Wrap(this.TableInfoList, USER_ROLE_GUEST)
	routeMap["/api/install/create/table"] = this.Wrap(this.CreateTable, USER_ROLE_GUEST)
	routeMap["/api/install/admin/list"] = this.Wrap(this.AdminList, USER_ROLE_GUEST)
	routeMap["/api/install/create/admin"] = this.Wrap(this.CreateAdmin, USER_ROLE_GUEST)
	routeMap["/api/install/validate/admin"] = this.Wrap(this.ValidateAdmin, USER_ROLE_GUEST)
	routeMap["/api/install/finish"] = this.Wrap(this.Finish, USER_ROLE_GUEST)

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

	db.LogMode(true)

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
func (this *InstallController) getTableMeta(gormDb *gorm.DB, entity IBase) (bool, []*gorm.StructField, []*gorm.StructField) {

	//挣扎一下，尝试获取建表语句。
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

//根据表名获取建表SQL语句
func (this *InstallController) getTableMetaList(db *gorm.DB) []*InstallTableInfo {

	var tableNames = []IBase{&Dashboard{}, &DownloadToken{}, &Footprint{}, &ImageCache{}, &Matter{}, &Preference{}, &Session{}, UploadToken{}, &User{}}
	var installTableInfos []*InstallTableInfo

	for _, iBase := range tableNames {
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

//验证表结构是否完整。会直接抛出异常
func (this *InstallController) validateTableMetaList(tableInfoList []*InstallTableInfo) {

	for _, tableInfo := range tableInfoList {
		if tableInfo.TableExist {
			if len(tableInfo.MissingFields) != 0 {

				var strs []string
				for _, v := range tableInfo.MissingFields {
					strs = append(strs, v.DBName)
				}

				this.PanicBadRequest(fmt.Sprintf("%s 表的以下字段缺失：%v", tableInfo.Name, strs))
			}
		} else {
			this.PanicBadRequest(tableInfo.Name + "表不存在")
		}
	}

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
func (this *InstallController) TableInfoList(writer http.ResponseWriter, request *http.Request) *WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	return this.Success(this.getTableMetaList(db))
}

//创建缺失数据库和表
func (this *InstallController) CreateTable(writer http.ResponseWriter, request *http.Request) *WebResult {

	var tableNames = []IBase{&Dashboard{}, &DownloadToken{}, &Footprint{}, &ImageCache{}, &Matter{}, &Preference{}, &Session{}, UploadToken{}, &User{}}
	var installTableInfos []*InstallTableInfo

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	for _, iBase := range tableNames {

		//补全缺失字段或者创建数据库表
		db1 := db.AutoMigrate(iBase)
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

//获取管理员列表(10条记录)
func (this *InstallController) AdminList(writer http.ResponseWriter, request *http.Request) *WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	var wp = &WherePair{}

	wp = wp.And(&WherePair{Query: "role = ?", Args: []interface{}{USER_ROLE_ADMINISTRATOR}})

	var users []*User
	db = db.Where(wp.Query, wp.Args...).Offset(0).Limit(10).Find(&users)

	this.PanicError(db.Error)

	return this.Success(users)
}

//创建管理员
func (this *InstallController) CreateAdmin(writer http.ResponseWriter, request *http.Request) *WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	adminUsername := request.FormValue("adminUsername")
	adminEmail := request.FormValue("adminEmail")
	adminPassword := request.FormValue("adminPassword")

	//验证超级管理员的信息
	if m, _ := regexp.MatchString(`^[0-9a-zA-Z_]+$`, adminUsername); !m {
		this.PanicBadRequest(`超级管理员用户名必填，且只能包含字母，数字和'_''`)
	}

	if len(adminPassword) < 6 {
		this.PanicBadRequest(`超级管理员密码长度至少为6位`)
	}

	if adminEmail == "" {
		this.PanicBadRequest(`超级管理员邮箱必填`)
	}

	//检查是否有重复。
	var count1 int64
	db1 := db.Model(&User{}).Where("email = ?", adminEmail).Count(&count1)
	this.PanicError(db1.Error)
	if count1 > 0 {
		this.PanicBadRequest(`该邮箱已存在`)
	}

	var count2 int64
	db2 := db.Model(&User{}).Where("username = ?", adminUsername).Count(&count2)
	this.PanicError(db2.Error)
	if count2 > 0 {
		this.PanicBadRequest(`该用户名已存在`)
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
	user.Password = GetBcrypt(adminPassword)
	user.Email = adminEmail
	user.Phone = ""
	user.Gender = USER_GENDER_UNKNOWN
	user.SizeLimit = -1
	user.Status = USER_STATUS_OK

	db3 := db.Create(user)
	this.PanicError(db3.Error)

	return this.Success("OK")

}

//(如果数据库中本身存在管理员了)验证管理员
func (this *InstallController) ValidateAdmin(writer http.ResponseWriter, request *http.Request) *WebResult {

	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	adminEmail := request.FormValue("adminEmail")
	adminPassword := request.FormValue("adminPassword")

	//验证超级管理员的信息
	if adminEmail == "" {
		this.PanicBadRequest(`超级管理员邮箱必填`)
	}
	if len(adminPassword) < 6 {
		this.PanicBadRequest(`超级管理员密码长度至少为6位`)
	}

	var existEmailUser = &User{}
	db = db.Where(&User{Email: adminEmail}).First(existEmailUser)
	if db.Error != nil {
		this.PanicBadRequest(fmt.Sprintf("%s对应的用户不存在", adminEmail))
	}

	if !MatchBcrypt(adminPassword, existEmailUser.Password) {
		this.PanicBadRequest("邮箱或密码错误")
	}

	if existEmailUser.Role != USER_ROLE_ADMINISTRATOR {
		this.PanicBadRequest("该账号不是管理员")
	}

	return this.Success("OK")

}

//完成系统安装
func (this *InstallController) Finish(writer http.ResponseWriter, request *http.Request) *WebResult {

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

	//要求数据库连接通畅
	db := this.openDbConnection(writer, request)
	defer this.closeDbConnection(db)

	//要求数据库完整。
	tableMetaList := this.getTableMetaList(db)
	this.validateTableMetaList(tableMetaList)

	//要求至少有一名管理员。
	var count1 int64
	db1 := db.Model(&User{}).Where("role = ?", USER_ROLE_ADMINISTRATOR).Count(&count1)
	this.PanicError(db1.Error)
	if count1 == 0 {
		this.PanicBadRequest(`请至少配置一名管理员`)
	}

	var configItem = &ConfigItem{
		//默认监听端口号
		ServerPort: CONFIG.ServerPort,
		//上传的文件路径，要求不以/结尾。如果没有指定，默认在根目录下的matter文件夹中。eg: /var/www/matter
		MatterPath: CONFIG.MatterPath,
		//mysql相关配置。
		//数据库端口
		MysqlPort: mysqlPort,
		//数据库Host
		MysqlHost: mysqlHost,
		//数据库名字
		MysqlSchema: mysqlSchema,
		//用户名
		MysqlUsername: mysqlUsername,
		//密码
		MysqlPassword: mysqlPassword,
	}

	//用json的方式输出返回值。为了让格式更好看。
	jsonStr, _ := jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent(configItem, "", " ")

	//写入到配置文件中（不能使用os.O_APPEND 否则会追加）
	filePath := GetConfPath() + "/tank.json"
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0777)
	this.PanicError(err)
	_, err = f.Write(jsonStr)
	this.PanicError(err)
	err = f.Close()
	this.PanicError(err)

	//通知配置文件安装完毕。
	CONFIG.InstallOk()

	//通知全局上下文，说系统安装好了
	CONTEXT.InstallOk()

	return this.Success("OK")
}
