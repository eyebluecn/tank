package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	uuid "github.com/nu7hatch/gouuid"
	"net/http"
	"time"
)

//@Service
type UserService struct {
	BaseBean
	userDao    *UserDao
	sessionDao *SessionDao

	//操作文件的锁。
	locker *cache.Table
}

func (this *UserService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.sessionDao)
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
		this.logger.Error("error while get cache" + err.Error())
	}

	//当前被锁住了。
	if cacheItem != nil && cacheItem.Data() != nil {
		panic(result.BadRequest("file is being operating, retry later"))
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
//1. 支持cookie形式 2.支持入参传入username和password 3.支持Basic Auth
func (this *UserService) PreHandle(writer http.ResponseWriter, request *http.Request) {

	//登录身份有效期以数据库中记录的为准

	//验证用户是否已经登录。
	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)

	if sessionId != "" {

		//去缓存中捞取
		cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
		if err != nil {
			this.logger.Error("获取缓存时出错了" + err.Error())
		}

		//缓存中没有，尝试去数据库捞取
		if cacheItem == nil || cacheItem.Data() == nil {
			session := this.sessionDao.FindByUuid(sessionId)
			if session != nil {
				duration := session.ExpireTime.Sub(time.Now())
				if duration <= 0 {
					this.logger.Error("登录信息已过期")
				} else {
					user := this.userDao.FindByUuid(session.UserUuid)
					if user != nil {
						//将用户装填进缓存中
						core.CONTEXT.GetSessionCache().Add(sessionId, duration, user)
					} else {
						this.logger.Error("没有找到对应的user %s", session.UserUuid)
					}
				}
			}
		}
	}

	//再尝试读取一次，这次从 USERNAME_KEY PASSWORD_KEY 中装填用户登录信息
	cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
	if err != nil {
		this.logger.Error("获取缓存时出错了" + err.Error())
	}

	if cacheItem == nil || cacheItem.Data() == nil {
		username := request.FormValue(core.USERNAME_KEY)
		password := request.FormValue(core.PASSWORD_KEY)

		if username == "" || password == "" {
			username, password, _ = request.BasicAuth()
		}

		if username != "" && password != "" {

			user := this.userDao.FindByUsername(username)
			if user == nil {
				this.logger.Error("%s 用户名或密码错误", core.USERNAME_KEY)
			} else {

				if !util.MatchBcrypt(password, user.Password) {
					this.logger.Error("%s 用户名或密码错误", core.USERNAME_KEY)
				} else {
					//装填一个临时的session用作后续使用。
					this.logger.Info("准备装载一个临时的用作。")
					timeUUID, _ := uuid.NewV4()
					uuidStr := string(timeUUID.String())
					request.Form[core.COOKIE_AUTH_KEY] = []string{uuidStr}

					//将用户装填进缓存中
					core.CONTEXT.GetSessionCache().Add(uuidStr, 10*time.Second, user)
				}
			}

		}
	}

}
