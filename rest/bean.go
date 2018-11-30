package rest

import (
	"net/http"
)

type IBean interface {
	Init()
	PanicError(err error);
	PanicWebError(msg string, code int);
}

type Bean struct {
	logger *Logger
}

func (this *Bean) Init() {
	this.logger = LOGGER
}

//处理错误的统一方法
func (this *Bean) PanicError(err error) {
	if err != nil {
		panic(&WebError{Msg: err.Error(), Code: http.StatusInternalServerError})
	}
}

//处理错误的统一方法
func (this *Bean) PanicWebError(msg string, httpStatusCode int) {
	panic(&WebError{Msg: msg, Code: httpStatusCode})
}

//能找到一个user就找到一个
func (this *Bean) findUser(writer http.ResponseWriter, request *http.Request) *User {

	//验证用户是否已经登录。
	sessionCookie, err := request.Cookie(COOKIE_AUTH_KEY)
	if err != nil {
		this.logger.Warn("cookie 信息不存在~")
		return nil
	}

	sessionId := sessionCookie.Value

	//去缓存中捞取看看
	cacheItem, err := CONTEXT.SessionCache.Value(sessionId)
	if err != nil {
		this.logger.Warn("获取缓存时出错了" + err.Error())
		return nil
	}

	if cacheItem == nil || cacheItem.Data() == nil {
		this.logger.Warn("cache item中已经不存在了 ")
		return nil
	}

	if value, ok := cacheItem.Data().(*User); ok {
		return value
	} else {
		this.logger.Error("cache item中的类型不是*User ")
	}

	return nil
}

//获取当前登录的用户，找不到就返回登录错误
func (this *Bean) checkUser(writer http.ResponseWriter, request *http.Request) *User {
	if this.findUser(writer, request) == nil {
		panic(ConstWebResult(CODE_WRAPPER_LOGIN))
	} else {
		return this.findUser(writer, request)
	}
}
