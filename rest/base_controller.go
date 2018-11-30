package rest

import (
	"fmt"
	"github.com/json-iterator/go"
	"go/types"
	"net/http"
)

type IController interface {
	IBean
	//注册自己固定的路由。
	RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request)
	//处理一些特殊的路由。
	HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool)
}
type BaseController struct {
	Bean
	userDao    *UserDao
	sessionDao *SessionDao
}

func (this *BaseController) Init() {

	this.Bean.Init()

	//手动装填本实例的Bean.
	b := CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = CONTEXT.GetBean(this.sessionDao)
	if b, ok := b.(*SessionDao); ok {
		this.sessionDao = b
	}

}

//注册自己的路由。
func (this *BaseController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {
	//每个Controller需要主动注册自己的路由。
	return make(map[string]func(writer http.ResponseWriter, request *http.Request))
}

//处理一些特殊的接口，比如参数包含在路径中,一般情况下，controller不将参数放在url路径中
func (this *BaseController) HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool) {
	return nil, false
}

//需要进行登录验证的wrap包装
func (this *BaseController) Wrap(f func(writer http.ResponseWriter, request *http.Request) *WebResult, qualifiedRole string) func(w http.ResponseWriter, r *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		//writer和request赋值给自己。

		var webResult *WebResult = nil

		//只有游客接口不需要登录
		if qualifiedRole != USER_ROLE_GUEST {
			user := this.checkUser(writer, request)

			if user.Status == USER_STATUS_DISABLED {
				//判断用户是否被禁用。
				webResult = ConstWebResult(CODE_WRAPPER_USER_DISABLED)
			} else {
				if qualifiedRole == USER_ROLE_ADMINISTRATOR && user.Role != USER_ROLE_ADMINISTRATOR {
					webResult = ConstWebResult(CODE_WRAPPER_UNAUTHORIZED)
				} else {
					webResult = f(writer, request)
				}
			}

		} else {
			webResult = f(writer, request)
		}

		//输出的是json格式
		if webResult != nil {
			//返回的内容申明是json，utf-8
			writer.Header().Set("Content-Type", "application/json;charset=UTF-8")

			//用json的方式输出返回值。
			var json = jsoniter.ConfigCompatibleWithStandardLibrary
			b, err := json.Marshal(webResult)

			this.PanicError(err)

			writer.WriteHeader(FetchHttpStatus(webResult.Code))

			fmt.Fprintf(writer, string(b))
		} else {
			//输出的内容是二进制的。

		}

	}
}

//返回成功的结果。
func (this *BaseController) Success(data interface{}) *WebResult {
	var webResult *WebResult = nil
	if value, ok := data.(string); ok {
		webResult = &WebResult{Code: CODE_WRAPPER_OK.Code, Msg: value}
	} else if value, ok := data.(*WebResult); ok {
		webResult = value
	} else if _, ok := data.(types.Nil); ok {
		webResult = ConstWebResult(CODE_WRAPPER_OK)
	} else {
		webResult = &WebResult{Code: CODE_WRAPPER_OK.Code, Data: data}
	}
	return webResult
}

//返回错误的结果。
func (this *BaseController) Error(err interface{}) *WebResult {
	var webResult *WebResult = nil
	if value, ok := err.(string); ok {
		webResult = &WebResult{Code: CODE_WRAPPER_UNKNOWN.Code, Msg: value}
	} else if _, ok := err.(int); ok {
		webResult = ConstWebResult(CODE_WRAPPER_UNKNOWN)
	} else if value, ok := err.(*WebResult); ok {
		webResult = value
	} else if value, ok := err.(error); ok {
		webResult = &WebResult{Code: CODE_WRAPPER_UNKNOWN.Code, Msg: value.Error()}
	} else {
		webResult = &WebResult{Code: CODE_WRAPPER_UNKNOWN.Code, Msg: "服务器未知错误"}
	}
	return webResult
}

//能找到一个user就找到一个
func (this *BaseController) findUser(writer http.ResponseWriter, request *http.Request) *User {

	//验证用户是否已经登录。
	sessionCookie, err := request.Cookie(COOKIE_AUTH_KEY)
	if err != nil {
		this.logger.Warn("获取用户cookie信息失败啦~")
		return nil
	}

	sessionId := sessionCookie.Value

	this.logger.Info("findUser sessionId = %s", sessionId)

	//去缓存中捞取看看
	cacheItem, err := CONTEXT.SessionCache.Value(sessionId)
	if err != nil {
		this.logger.Warn("获取缓存时出错了" + err.Error())
		return nil
	}

	if cacheItem.Data() == nil {
		this.logger.Warn("cache item中已经不存在了 " + err.Error())
		return nil
	}

	if value, ok := cacheItem.Data().(*User); ok {
		return value
	} else {
		this.logger.Error("cache item中的类型不是*User ")
	}

	return nil
}

func (this *BaseController) checkUser(writer http.ResponseWriter, request *http.Request) *User {
	if this.findUser(writer, request) == nil {
		panic(ConstWebResult(CODE_WRAPPER_LOGIN))
	} else {
		return this.findUser(writer, request)
	}
}

//允许跨域请求
func (this *BaseController) allowCORS(writer http.ResponseWriter) {
	writer.Header().Add("Access-Control-Allow-Origin", "*")
	writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	writer.Header().Add("Access-Control-Max-Age", "3600")
}
