package support

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/rest"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
)

type TankContext struct {
	//db connection
	db *gorm.DB
	//session cache
	SessionCache *cache.Table
	//bean map.
	BeanMap map[string]core.Bean
	//controller map
	ControllerMap map[string]core.Controller
	//router
	Router *TankRouter
}

func (this *TankContext) Init() {

	//create session cache
	this.SessionCache = cache.NewTable()

	//init map
	this.BeanMap = make(map[string]core.Bean)
	this.ControllerMap = make(map[string]core.Controller)

	//register beans. This method will put Controllers to ControllerMap.
	this.registerBeans()

	//init every bean.
	this.initBeans()

	//create and init router.
	this.Router = NewRouter()

	//if the application is installed. Bean's Bootstrap method will be invoked.
	this.InstallOk()

}

func (this *TankContext) GetDB() *gorm.DB {
	return this.db
}

func (this *TankContext) GetSessionCache() *cache.Table {
	return this.SessionCache
}

func (this *TankContext) GetControllerMap() map[string]core.Controller {
	return this.ControllerMap
}

func (this *TankContext) Cleanup() {
	for _, bean := range this.BeanMap {
		bean.Cleanup()
	}
}

//can serve as http server.
func (this *TankContext) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	this.Router.ServeHTTP(writer, request)
}

func (this *TankContext) OpenDb() {

	//log strategy.
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // slow SQL 1s
			LogLevel:                  logger.Silent, // log level. open when debug.
			IgnoreRecordNotFoundError: true,          // ignore ErrRecordNotFound
			Colorful:                  false,         // colorful print
		},
	)

	//table name strategy.
	namingStrategy := core.CONFIG.NamingStrategy()

	if core.CONFIG.DbType() == "sqlite" {

		var err error = nil
		this.db, err = gorm.Open(sqlite.Open(core.CONFIG.SqliteFolder()+"/tank.sqlite"), &gorm.Config{Logger: dbLogger, NamingStrategy: namingStrategy})

		if err != nil {
			core.LOGGER.Panic("failed to connect mysql database")
		}

		//sqlite lock issue. https://gist.github.com/mrnugget/0eda3b2b53a70fa4a894
		phyDb, err := this.db.DB()
		phyDb.SetMaxOpenConns(1)

	} else {

		var err error = nil
		this.db, err = gorm.Open(mysql.Open(core.CONFIG.MysqlUrl()), &gorm.Config{Logger: dbLogger, NamingStrategy: namingStrategy})

		if err != nil {
			core.LOGGER.Panic("failed to connect mysql database")
		}
	}

}

func (this *TankContext) CloseDb() {

	if this.db != nil {
		db, err := this.db.DB()
		if err != nil {
			core.LOGGER.Error("occur error when get *sql.DB %s", err.Error())
		}
		err = db.Close()
		if err != nil {
			core.LOGGER.Error("occur error when closing db %s", err.Error())
		}
	}
}

func (this *TankContext) registerBean(bean core.Bean) {

	typeOf := reflect.TypeOf(bean)
	typeName := typeOf.String()

	if element, ok := bean.(core.Bean); ok {

		if _, ok := this.BeanMap[typeName]; ok {
			core.LOGGER.Error("%s has been registerd, skip", typeName)
		} else {
			this.BeanMap[typeName] = element

			//if is controller type, put into ControllerMap
			if controller, ok1 := bean.(core.Controller); ok1 {
				this.ControllerMap[typeName] = controller
			}

		}

	} else {
		core.LOGGER.Panic("%s is not the Bean type", typeName)
	}

}

func (this *TankContext) registerBeans() {

	//alien
	this.registerBean(new(rest.AlienController))
	this.registerBean(new(rest.AlienService))

	//bridge
	this.registerBean(new(rest.BridgeDao))
	this.registerBean(new(rest.BridgeService))

	//dashboard
	this.registerBean(new(rest.DashboardController))
	this.registerBean(new(rest.DashboardDao))
	this.registerBean(new(rest.DashboardService))

	//downloadToken
	this.registerBean(new(rest.DownloadTokenDao))

	//imageCache
	this.registerBean(new(rest.ImageCacheController))
	this.registerBean(new(rest.ImageCacheDao))
	this.registerBean(new(rest.ImageCacheService))

	//install
	this.registerBean(new(rest.InstallController))

	//matter
	this.registerBean(new(rest.MatterController))
	this.registerBean(new(rest.MatterDao))
	this.registerBean(new(rest.MatterService))

	//preference
	this.registerBean(new(rest.PreferenceController))
	this.registerBean(new(rest.PreferenceDao))
	this.registerBean(new(rest.PreferenceService))

	//footprint
	this.registerBean(new(rest.FootprintController))
	this.registerBean(new(rest.FootprintDao))
	this.registerBean(new(rest.FootprintService))

	//session
	this.registerBean(new(rest.SessionDao))
	this.registerBean(new(rest.SessionService))

	//share
	this.registerBean(new(rest.ShareController))
	this.registerBean(new(rest.ShareDao))
	this.registerBean(new(rest.ShareService))

	//uploadToken
	this.registerBean(new(rest.UploadTokenDao))

	//task
	this.registerBean(new(rest.TaskService))

	//user
	this.registerBean(new(rest.UserController))
	this.registerBean(new(rest.UserDao))
	this.registerBean(new(rest.UserService))

	//webdav
	this.registerBean(new(rest.DavController))
	this.registerBean(new(rest.DavService))

}

func (this *TankContext) GetBean(bean core.Bean) core.Bean {

	typeOf := reflect.TypeOf(bean)
	typeName := typeOf.String()

	if val, ok := this.BeanMap[typeName]; ok {
		return val
	} else {
		core.LOGGER.Panic("%s not registered", typeName)
		return nil
	}
}

func (this *TankContext) initBeans() {

	for _, bean := range this.BeanMap {
		bean.Init()
	}
}

//if application installed. invoke this method.
func (this *TankContext) InstallOk() {

	if core.CONFIG.Installed() {
		this.OpenDb()

		for _, bean := range this.BeanMap {
			bean.Bootstrap()
		}
	}

}

func (this *TankContext) Destroy() {
	this.CloseDb()
}
