package rest

import (
	"github.com/eyebluecn/tank/code/config"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"time"
)

//@Service
type UserService struct {
	Bean
	userDao    *UserDao
	sessionDao *SessionDao

	//操作文件的锁。
	locker *cache.Table
}

//初始化方法
func (this *UserService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = CONTEXT.GetBean(this.sessionDao)
	if b, ok := b.(*SessionDao); ok {
		this.sessionDao = b
	}

	//创建一个用于存储用户文件锁的缓存。
	this.locker = cache.NewTable()
}

//对某个用户进行加锁。加锁阶段用户是不允许操作文件的。
func (this *UserService) MatterLock(userUuid string) {
	//如果已经是锁住的状态，直接报错

	//去缓存中捞取
	cacheItem, err := this.locker.Value(userUuid)
	if err != nil {
		this.logger.Error("获取缓存时出错了" + err.Error())
	}

	//当前被锁住了。
	if cacheItem != nil && cacheItem.Data() != nil {
		panic(result.BadRequest("当前正在进行文件操作，请稍后再试！"))
	}

	//添加一把新锁，有效期为12小时
	duration := 12 * time.Hour
	this.locker.Add(userUuid, duration, true)
}

//对某个用户解锁，解锁后用户可以操作文件。
func (this *UserService) MatterUnlock(userUuid string) {

	exist := this.locker.Exists(userUuid)
	if exist {
		_, err := this.locker.Delete(userUuid)
		this.PanicError(err)
	} else {
		this.logger.Error("%s已经不存在matter锁了，解锁错误。", userUuid)
	}
}

//装载session信息，如果session没有了根据cookie去装填用户信息。
//在所有的路由最初会调用这个方法
func (this *UserService) preHandle(writer http.ResponseWriter, request *http.Request) {

	//登录身份有效期以数据库中记录的为准

	//验证用户是否已经登录。
	sessionCookie, err := request.Cookie(config.COOKIE_AUTH_KEY)
	if err != nil {
		return
	}

	sessionId := sessionCookie.Value

	//去缓存中捞取
	cacheItem, err := CONTEXT.SessionCache.Value(sessionId)
	if err != nil {
		this.logger.Error("获取缓存时出错了" + err.Error())
	}

	//缓存中没有，尝试去数据库捞取
	if cacheItem == nil || cacheItem.Data() == nil {
		session := this.sessionDao.FindByUuid(sessionCookie.Value)
		if session != nil {
			duration := session.ExpireTime.Sub(time.Now())
			if duration <= 0 {
				this.logger.Error("登录信息已过期")
			} else {
				user := this.userDao.FindByUuid(session.UserUuid)
				if user != nil {
					//将用户装填进缓存中
					CONTEXT.SessionCache.Add(sessionCookie.Value, duration, user)
				} else {
					this.logger.Error("没有找到对应的user " + session.UserUuid)
				}
			}
		}
	}

}
