package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
)

type BaseBean struct {
	logger core.Logger
}

func (this *BaseBean) Init() {
	this.logger = core.LOGGER
}

func (this *BaseBean) Bootstrap() {

}

//clean up the application.
func (this *BaseBean) Cleanup() {

}

//shortcut for panic check.
func (this *BaseBean) PanicError(err error) {
	core.PanicError(err)
}

//find the current user from request.
func (this *BaseBean) findUser(request *http.Request) *User {

	//try to find from SessionCache.
	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)
	if sessionId == "" {
		return nil
	}

	cacheItem, err := core.CONTEXT.GetSessionCache().Value(sessionId)
	if err != nil {
		this.logger.Warn("error while get from session cache. sessionId = %s, error = %v", sessionId, err)
		return nil
	}

	if cacheItem == nil || cacheItem.Data() == nil {

		this.logger.Warn("cache item doesn't exist with sessionId = %s", sessionId)
		return nil
	}

	if value, ok := cacheItem.Data().(*User); ok {
		return value
	} else {
		this.logger.Error("cache item not store the *User")
	}

	return nil

}

//find current error. If not found, panic the LOGIN error.
func (this *BaseBean) checkUser(request *http.Request) *User {
	if this.findUser(request) == nil {
		panic(result.LOGIN)
	} else {
		return this.findUser(request)
	}
}
