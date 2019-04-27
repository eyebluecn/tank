package support

import (
	"fmt"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/rest"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/jinzhu/gorm"
	"net/http"
	"reflect"
)

//上下文，管理数据库连接，管理所有路由请求，管理所有的单例component.
type TankContext struct {
	//数据库连接
	db *gorm.DB
	//session缓存
	SessionCache *cache.Table
	//各类的Bean Map。这里面是包含ControllerMap中所有元素
	BeanMap map[string]core.Bean
	//只包含了Controller的map
	ControllerMap map[string]core.Controller
	//处理所有路由请求
	Router *TankRouter
}

//初始化上下文
func (this *TankContext) Init() {

	//创建一个用于存储session的缓存。
	this.SessionCache = cache.NewTable()

	//初始化Map
	this.BeanMap = make(map[string]core.Bean)
	this.ControllerMap = make(map[string]core.Controller)

	//注册各类Beans.在这个方法里面顺便把Controller装入ControllerMap中去。
	this.registerBeans()

	//初始化每个bean.
	this.initBeans()

	//初始化Router. 这个方法要在Bean注册好了之后才能。
	this.Router = NewRouter()

	//如果数据库信息配置好了，就直接打开数据库连接 同时执行Bean的ConfigPost方法
	this.InstallOk()

}

//获取数据库对象
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

//响应http的能力
func (this *TankContext) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	this.Router.ServeHTTP(writer, request)
}

func (this *TankContext) OpenDb() {

	var err error = nil
	this.db, err = gorm.Open("mysql", core.CONFIG.MysqlUrl())

	if err != nil {
		core.LOGGER.Panic("failed to connect mysql database")
	}

	//是否打开sql日志(在调试阶段可以打开，以方便查看执行的SQL)
	this.db.LogMode(false)
}

func (this *TankContext) CloseDb() {

	if this.db != nil {
		err := this.db.Close()
		if err != nil {
			core.LOGGER.Error("关闭数据库连接出错 %s", err.Error())
		}
	}
}

//注册一个Bean
func (this *TankContext) registerBean(bean core.Bean) {

	typeOf := reflect.TypeOf(bean)
	typeName := typeOf.String()

	if element, ok := bean.(core.Bean); ok {

		err := fmt.Sprintf("【%s】已经被注册了，跳过。", typeName)
		if _, ok := this.BeanMap[typeName]; ok {
			core.LOGGER.Error(fmt.Sprintf(err))
		} else {
			this.BeanMap[typeName] = element

			//看看是不是controller类型，如果是，那么单独放在ControllerMap中。
			if controller, ok1 := bean.(core.Controller); ok1 {
				this.ControllerMap[typeName] = controller
			}

		}

	} else {
		core.LOGGER.Panic("注册的【%s】不是Bean类型。", typeName)
	}

}

//注册各个Beans
func (this *TankContext) registerBeans() {

	//alien
	this.registerBean(new(rest.AlienController))
	this.registerBean(new(rest.AlienService))

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

	//uploadToken
	this.registerBean(new(rest.UploadTokenDao))

	//user
	this.registerBean(new(rest.UserController))
	this.registerBean(new(rest.UserDao))
	this.registerBean(new(rest.UserService))

	//webdav
	this.registerBean(new(rest.DavController))
	this.registerBean(new(rest.DavService))

}

//从Map中获取某个Bean.
func (this *TankContext) GetBean(bean core.Bean) core.Bean {

	typeOf := reflect.TypeOf(bean)
	typeName := typeOf.String()

	if val, ok := this.BeanMap[typeName]; ok {
		return val
	} else {
		core.LOGGER.Panic("【%s】没有注册。", typeName)
		return nil
	}
}

//初始化每个Bean
func (this *TankContext) initBeans() {

	for _, bean := range this.BeanMap {
		bean.Init()
	}
}

//系统如果安装好了就调用这个方法。
func (this *TankContext) InstallOk() {

	if core.CONFIG.Installed() {
		this.OpenDb()

		for _, bean := range this.BeanMap {
			bean.Bootstrap()
		}
	}

}

//销毁的方法
func (this *TankContext) Destroy() {
	this.CloseDb()
}
