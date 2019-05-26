package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	gouuid "github.com/nu7hatch/gouuid"
	"net/http"
	"time"
)

//@Service
type UserService struct {
	BaseBean
	userDao    *UserDao
	sessionDao *SessionDao

	//file lock
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

	//create a lock cache.
	this.locker = cache.NewTable()
}

//lock a user's operation. If lock, user cannot operate file.
func (this *UserService) MatterLock(userUuid string) {

	cacheItem, err := this.locker.Value(userUuid)
	if err != nil {
		this.logger.Error("error while get cache" + err.Error())
	}

	if cacheItem != nil && cacheItem.Data() != nil {
		panic(result.BadRequest("file is being operating, retry later"))
	}

	duration := 12 * time.Hour
	this.locker.Add(userUuid, duration, true)
}

//unlock
func (this *UserService) MatterUnlock(userUuid string) {

	exist := this.locker.Exists(userUuid)
	if exist {
		_, err := this.locker.Delete(userUuid)
		this.PanicError(err)
	} else {
		this.logger.Error("unlock error. %s has no matter lock ", userUuid)
	}
}

//load session to SessionCache. This method will be invoked in every request.
//authorize by 1. cookie 2. username and password in request form. 3. Basic Auth
func (this *UserService) PreHandle(writer http.ResponseWriter, request *http.Request) {

	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)

	if sessionId != "" {

		cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
		if err != nil {
			this.logger.Error("occur error will get session cache %s", err.Error())
		}

		//if no cache. try to find in db.
		if cacheItem == nil || cacheItem.Data() == nil {
			session := this.sessionDao.FindByUuid(sessionId)
			if session != nil {
				duration := session.ExpireTime.Sub(time.Now())
				if duration <= 0 {
					this.logger.Error("login info has expired.")
				} else {
					user := this.userDao.FindByUuid(session.UserUuid)
					if user != nil {
						core.CONTEXT.GetSessionCache().Add(sessionId, duration, user)
					} else {
						this.logger.Error("no user with sessionId %s", session.UserUuid)
					}
				}
			}
		}
	}

	//try to auth by USERNAME_KEY PASSWORD_KEY
	cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
	if err != nil {
		this.logger.Error("occur error will get session cache %s", err.Error())
	}

	if cacheItem == nil || cacheItem.Data() == nil {
		username := request.FormValue(core.USERNAME_KEY)
		password := request.FormValue(core.PASSWORD_KEY)

		//try to read from BasicAuth
		if username == "" || password == "" {
			username, password, _ = request.BasicAuth()
		}

		if username != "" && password != "" {

			user := this.userDao.FindByUsername(username)
			if user == nil {
				this.logger.Error("%s no such user in db.", username)
			} else {

				if !util.MatchBcrypt(password, user.Password) {
					this.logger.Error("%s password error", username)
				} else {

					this.logger.Info("load a temp session by username and password.")
					timeUUID, _ := gouuid.NewV4()
					uuidStr := string(timeUUID.String())
					request.Form[core.COOKIE_AUTH_KEY] = []string{uuidStr}

					core.CONTEXT.GetSessionCache().Add(uuidStr, 10*time.Second, user)
				}
			}

		}
	}

}

//find a cache user by its userUuid
func (this *UserService) FindCacheUsersByUuid(userUuid string) []*User {

	var users []*User
	//let session user work.
	core.CONTEXT.GetSessionCache().Foreach(func(key interface{}, cacheItem *cache.Item) {
		if cacheItem == nil || cacheItem.Data() == nil {
			return
		}
		if value, ok := cacheItem.Data().(*User); ok {
			var user = value
			if user.Uuid == userUuid {
				users = append(users, user)
			}
		} else {
			this.logger.Error("cache item not store the *User")
		}
	})

	return users
}
