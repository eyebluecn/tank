package rest

import (
	"sync"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/eyebluecn/tank/code/tool/uuid"
	"net/http"
	"os"
)

// @Service
type UserService struct {
	BaseBean
	userDao    *UserDao
	sessionDao *SessionDao

	spaceService *SpaceService

	//file lock (legacy, kept for compatibility)
	locker *cache.Table
	// matterLocker is used for atomic matter operations with retry support
	matterLocker sync.Map

	matterDao        *MatterDao
	matterService    *MatterService
	imageCacheDao    *ImageCacheDao
	spaceDao         *SpaceDao
	spaceMemberDao   *SpaceMemberDao
	shareDao         *ShareDao
	shareService     *ShareService
	downloadTokenDao *DownloadTokenDao
	uploadTokenDao   *UploadTokenDao
	footprintDao     *FootprintDao
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

	b = core.CONTEXT.GetBean(this.spaceService)
	if b, ok := b.(*SpaceService); ok {
		this.spaceService = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}
	b = core.CONTEXT.GetBean(this.shareService)
	if b, ok := b.(*ShareService); ok {
		this.shareService = b
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if b, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadTokenDao)
	if b, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = b
	}

	b = core.CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	//create a lock cache.
	this.locker = cache.NewTable()
}

// lock a user's operation with retry mechanism.
// This allows concurrent operations to wait for each other instead of immediately failing.
// Uses sync.Map for thread-safe atomic operations.
func (this *UserService) MatterLock(userUuid string) {
	maxRetries := 30        // Maximum retry attempts (30 * 200ms = 6 seconds max wait)
	retryInterval := 200    // Retry interval in milliseconds

	for i := 0; i < maxRetries; i++ {
		// Try to acquire lock atomically
		if _, loaded := this.matterLocker.LoadOrStore(userUuid, true); !loaded {
			// Successfully acquired the lock
			return
		}

		// Lock is held by another operation, wait and retry
		if i < maxRetries-1 {
			time.Sleep(time.Duration(retryInterval) * time.Millisecond)
		}
	}

	// After all retries, still cannot acquire lock
	panic(result.BadRequest("file is being operating, retry later"))
}

// unlock
func (this *UserService) MatterUnlock(userUuid string) {
	this.matterLocker.Delete(userUuid)
}

// load session to SessionCache. This method will be invoked in every request.
// authorize by 1. cookie 2. username and password in request form. 3. Basic Auth
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
					timeUUID, _ := uuid.NewV4()
					uuidStr := string(timeUUID.String())
					request.Form[core.COOKIE_AUTH_KEY] = []string{uuidStr}

					core.CONTEXT.GetSessionCache().Add(uuidStr, 10*time.Second, user)
				}
			}

		}
	}

}

// find a cache user by its userUuid
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

// remove cache user by its userUuid
func (this *UserService) RemoveCacheUserByUuid(userUuid string) {

	var sessionId interface{}
	//let session user work.
	core.CONTEXT.GetSessionCache().Foreach(func(key interface{}, cacheItem *cache.Item) {
		if cacheItem == nil || cacheItem.Data() == nil {
			return
		}
		if value, ok := cacheItem.Data().(*User); ok {
			var user = value
			if user.Uuid == userUuid {
				sessionId = key
				this.logger.Info("sessionId %v", key)
			}
		} else {
			this.logger.Error("cache item not store the *User")
		}
	})

	exists := core.CONTEXT.GetSessionCache().Exists(sessionId)
	if exists {
		_, err := core.CONTEXT.GetSessionCache().Delete(sessionId)
		if err != nil {
			this.logger.Error("occur error when deleting cache user.")
		}
	}
}

// create user
func (this *UserService) CreateUser(request *http.Request, username string, sizeLimit int64, totalSizeLimit int64, password string, role string) *User {

	user := &User{
		Username: username,
		Password: util.GetBcrypt(password),
		Role:     role,
		Status:   USER_STATUS_OK,
	}

	user = this.userDao.Create(user)

	//create space.
	space := this.spaceService.CreateSpace(request, username, user, sizeLimit, totalSizeLimit, SPACE_TYPE_PRIVATE)

	//update user's space.
	user.SpaceUuid = space.Uuid
	this.userDao.Save(user)

	user.Space = space

	return user

}

// delete user
func (this *UserService) DeleteUser(request *http.Request, currentUser *User) {

	//delete from cache
	this.logger.Info("delete from cache userUuid = %s", currentUser.Uuid)
	this.RemoveCacheUserByUuid(currentUser.Uuid)

	//delete download tokens
	this.logger.Info("delete download tokens")
	this.downloadTokenDao.DeleteByUserUuid(currentUser.Uuid)

	//delete upload tokens
	this.logger.Info("delete upload tokens")
	this.uploadTokenDao.DeleteByUserUuid(currentUser.Uuid)

	//delete footprints
	this.logger.Info("delete footprints")
	this.footprintDao.DeleteByUserUuid(currentUser.Uuid)

	//delete session
	this.logger.Info("delete session")
	this.sessionDao.DeleteByUserUuid(currentUser.Uuid)

	//delete shares and bridges
	this.logger.Info("elete shares and bridges")
	this.shareService.DeleteSharesByUser(request, currentUser)

	//delete caches
	this.logger.Info("delete caches")
	this.imageCacheDao.DeleteByUserUuid(currentUser.Uuid)

	//delete matters
	this.logger.Info("delete matters")
	this.matterDao.DeleteByUserUuid(currentUser.Uuid)

	//delete space members
	space := this.spaceDao.CheckByUuid(currentUser.SpaceUuid)
	this.logger.Info("delete space members")
	this.spaceMemberDao.DeleteBySpaceUuid(space.Uuid)

	//delete spaces
	this.logger.Info("delete spaces")
	this.spaceDao.DeleteByUserUuid(currentUser.Uuid)

	//delete this user
	this.logger.Info("delete this user.")
	this.userDao.Delete(currentUser)

	//delete files from disk.
	this.logger.Info("delete files from disk. %s", GetUserSpaceRootDir(currentUser.Username))
	err := os.RemoveAll(GetUserSpaceRootDir(currentUser.Username))
	this.PanicError(err)

}
