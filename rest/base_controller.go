package rest

import (
	"fmt"
	"github.com/json-iterator/go"
	"go/types"
	"net/http"
	"time"
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

func (this *BaseController) Init(context *Context) {

	this.Bean.Init(context)

	//手动装填本实例的Bean.
	b := context.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = context.GetBean(this.sessionDao)
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
func (this *BaseController) Wrap(f func(writer http.ResponseWriter, request *http.Request) *WebResult, role string) func(w http.ResponseWriter, r *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		//writer和request赋值给自己。

		var webResult *WebResult = nil

		//只有游客接口不需要登录
		if role != USER_ROLE_GUEST {
			user := this.checkUser(writer, request)
			if role == USER_ROLE_ADMINISTRATOR && user.Role != USER_ROLE_ADMINISTRATOR {
				webResult = ConstWebResult(RESULT_CODE_UNAUTHORIZED)
			} else {
				webResult = f(writer, request)
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
			b, _ := json.Marshal(webResult)

			if webResult.Code == RESULT_CODE_OK {
				writer.WriteHeader(http.StatusOK)
			} else {
				writer.WriteHeader(http.StatusBadRequest)
			}

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
		webResult = &WebResult{Code: RESULT_CODE_OK, Msg: value}
	} else if value, ok := data.(*WebResult); ok {
		webResult = value
	} else if _, ok := data.(types.Nil); ok {
		webResult = ConstWebResult(RESULT_CODE_OK)
	} else {
		webResult = &WebResult{Code: RESULT_CODE_OK, Data: data}
	}
	return webResult
}

//返回错误的结果。
func (this *BaseController) Error(err interface{}) *WebResult {
	var webResult *WebResult = nil
	if value, ok := err.(string); ok {
		webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: value}
	} else if value, ok := err.(int); ok {
		webResult = ConstWebResult(value)
	} else if value, ok := err.(*WebResult); ok {
		webResult = value
	} else if value, ok := err.(error); ok {
		webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: value.Error()}
	} else {
		webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: "服务器未知错误"}
	}
	return webResult
}

func (this *BaseController) checkLogin(writer http.ResponseWriter, request *http.Request) (*Session, *User) {

	//验证用户是否已经登录。
	sessionCookie, err := request.Cookie(COOKIE_AUTH_KEY)
	if err != nil {
		panic(ConstWebResult(RESULT_CODE_LOGIN))
	}

	session := this.sessionDao.FindByUuid(sessionCookie.Value)
	if session == nil {
		panic(ConstWebResult(RESULT_CODE_LOGIN))
	} else {
		if session.ExpireTime.Before(time.Now()) {
			panic(ConstWebResult(RESULT_CODE_LOGIN_EXPIRED))
		} else {

			user := this.userDao.FindByUuid(session.UserUuid)
			if user == nil {
				panic(ConstWebResult(RESULT_CODE_LOGIN_INVALID))
			} else {
				return session, user
			}

		}
	}

}

func (this *BaseController) checkUser(writer http.ResponseWriter, request *http.Request) *User {

	_, user := this.checkLogin(writer, request)
	return user

}

//允许跨域请求
func (this *BaseController) allowCORS(writer http.ResponseWriter) {
	writer.Header().Add("Access-Control-Allow-Origin", "*")
	writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	writer.Header().Add("Access-Control-Max-Age", "3600")
}
