package rest

import (
	"net/http"
)

type IBean interface {
	//初始化方法
	Init()
	//所有配置都加载完成后调用的方法，包括数据库加载完毕
	ConfigPost()
	//快速的Panic方法
	PanicError(err error)
}

type Bean struct {
	logger *Logger
}

func (this *Bean) Init() {
	this.logger = LOGGER
}

func (this *Bean) ConfigPost() {

}

//处理错误的统一方法 可以省去if err!=nil 这段代码
func (this *Bean) PanicError(err error) {
	if err != nil {
		panic(err)
	}
}

//请求参数有问题
func (this *Bean) PanicBadRequest(msg string) {
	panic(CustomWebResult(CODE_WRAPPER_BAD_REQUEST, msg))
}

//没有权限
func (this *Bean) PanicUnauthorized(msg string) {
	panic(CustomWebResult(CODE_WRAPPER_UNAUTHORIZED, msg))
}

//没有找到
func (this *Bean) PanicNotFound(msg string) {
	panic(CustomWebResult(CODE_WRAPPER_NOT_FOUND, msg))
}

//服务器内部出问题
func (this *Bean) PanicServer(msg string) {
	panic(CustomWebResult(CODE_WRAPPER_UNKNOWN, msg))
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
