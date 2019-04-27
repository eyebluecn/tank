package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/config"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/jinzhu/gorm"
	"reflect"
)

//全局唯一的上下文(在main函数中初始化)
var CONTEXT = &Context{}

//上下文，管理数据库连接，管理所有路由请求，管理所有的单例component.
type Context struct {
	//数据库连接
	DB *gorm.DB
	//session缓存
	SessionCache *cache.Table
	//各类的Bean Map。这里面是包含ControllerMap中所有元素
	BeanMap map[string]core.IBean
	//只包含了Controller的map
	ControllerMap map[string]IController
	//处理所有路由请求
	Router *Router
}

//初始化上下文
func (this *Context) Init() {

	//创建一个用于存储session的缓存。
	this.SessionCache = cache.NewTable()

	//初始化Map
	this.BeanMap = make(map[string]core.IBean)
	this.ControllerMap = make(map[string]IController)

	//注册各类Beans.在这个方法里面顺便把Controller装入ControllerMap中去。
	this.registerBeans()

	//初始化每个bean.
	this.initBeans()

	//初始化Router. 这个方法要在Bean注册好了之后才能。
	this.Router = NewRouter()

	//如果数据库信息配置好了，就直接打开数据库连接 同时执行Bean的ConfigPost方法
	this.InstallOk()

}

func (this *Context) OpenDb() {

	var err error = nil
	this.DB, err = gorm.Open("mysql", config.CONFIG.MysqlUrl)

	if err != nil {
		core.LOGGER.Panic("failed to connect mysql database")
	}

	//是否打开sql日志(在调试阶段可以打开，以方便查看执行的SQL)
	this.DB.LogMode(false)
}

func (this *Context) CloseDb() {

	if this.DB != nil {
		err := this.DB.Close()
		if err != nil {
			core.LOGGER.Error("关闭数据库连接出错 %s", err.Error())
		}
	}
}

//注册一个Bean
func (this *Context) registerBean(bean core.IBean) {

	typeOf := reflect.TypeOf(bean)
	typeName := typeOf.String()

	if element, ok := bean.(core.IBean); ok {

		err := fmt.Sprintf("【%s】已经被注册了，跳过。", typeName)
		if _, ok := this.BeanMap[typeName]; ok {
			core.LOGGER.Error(fmt.Sprintf(err))
		} else {
			this.BeanMap[typeName] = element

			//看看是不是controller类型，如果是，那么单独放在ControllerMap中。
			if controller, ok1 := bean.(IController); ok1 {
				this.ControllerMap[typeName] = controller
			}

		}

	} else {
		core.LOGGER.Panic("注册的【%s】不是Bean类型。", typeName)
	}

}

//注册各个Beans
func (this *Context) registerBeans() {

	//alien
	this.registerBean(new(AlienController))
	this.registerBean(new(AlienService))

	//dashboard
	this.registerBean(new(DashboardController))
	this.registerBean(new(DashboardDao))
	this.registerBean(new(DashboardService))

	//downloadToken
	this.registerBean(new(DownloadTokenDao))

	//imageCache
	this.registerBean(new(ImageCacheController))
	this.registerBean(new(ImageCacheDao))
	this.registerBean(new(ImageCacheService))

	//install
	this.registerBean(new(InstallController))

	//matter
	this.registerBean(new(MatterController))
	this.registerBean(new(MatterDao))
	this.registerBean(new(MatterService))

	//preference
	this.registerBean(new(PreferenceController))
	this.registerBean(new(PreferenceDao))
	this.registerBean(new(PreferenceService))

	//footprint
	this.registerBean(new(FootprintController))
	this.registerBean(new(FootprintDao))
	this.registerBean(new(FootprintService))

	//session
	this.registerBean(new(SessionDao))
	this.registerBean(new(SessionService))

	//uploadToken
	this.registerBean(new(UploadTokenDao))

	//user
	this.registerBean(new(UserController))
	this.registerBean(new(UserDao))
	this.registerBean(new(UserService))

	//webdav
	this.registerBean(new(DavController))
	this.registerBean(new(DavService))

}

//从Map中获取某个Bean.
func (this *Context) GetBean(bean core.IBean) core.IBean {

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
func (this *Context) initBeans() {

	for _, bean := range this.BeanMap {
		bean.Init()
	}
}

//系统如果安装好了就调用这个方法。
func (this *Context) InstallOk() {

	if config.CONFIG.Installed {
		this.OpenDb()

		for _, bean := range this.BeanMap {
			bean.Bootstrap()
		}
	}

}

//销毁的方法
func (this *Context) Destroy() {
	this.CloseDb()
}
