package rest

import (
	"fmt"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"os"
	"strings"
)

//用于处理所有前来的请求
type Router struct {
	context  *Context
	routeMap map[string]func(writer http.ResponseWriter, request *http.Request)
}

//构造方法
func NewRouter(context *Context) *Router {
	router := &Router{
		context:  context,
		routeMap: make(map[string]func(writer http.ResponseWriter, request *http.Request)),
	}

	for _, controller := range context.ControllerMap {
		routes := controller.RegisterRoutes()
		for k, v := range routes {
			router.routeMap[k] = v
		}
	}
	return router

}

//全局的异常捕获
func (this *Router) GlobalPanicHandler(writer http.ResponseWriter, request *http.Request) {
	if err := recover(); err != nil {

		LogError(fmt.Sprintf("全局异常: %v", err))

		var webResult *WebResult = nil
		if value, ok := err.(string); ok {
			writer.WriteHeader(http.StatusBadRequest)
			webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: value}
		} else if value, ok := err.(int); ok {
			writer.WriteHeader(http.StatusBadRequest)
			webResult = ConstWebResult(value)
		} else if value, ok := err.(*WebResult); ok {
			writer.WriteHeader(http.StatusBadRequest)
			webResult = value
		} else if value, ok := err.(WebResult); ok {
			writer.WriteHeader(http.StatusBadRequest)
			webResult = &value
		} else if value, ok := err.(*WebError); ok {
			writer.WriteHeader(value.Code)
			webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: value.Msg}
		} else if value, ok := err.(WebError); ok {
			writer.WriteHeader((&value).Code)
			webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: (&value).Msg}
		} else if value, ok := err.(error); ok {
			writer.WriteHeader(http.StatusBadRequest)
			webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: value.Error()}
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			webResult = &WebResult{Code: RESULT_CODE_UTIL_EXCEPTION, Msg: "服务器未知错误"}
		}

		//输出的是json格式 返回的内容申明是json，utf-8
		writer.Header().Set("Content-Type", "application/json;charset=UTF-8")

		//用json的方式输出返回值。
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		b, _ := json.Marshal(webResult)

		fmt.Fprintf(writer, string(b))
	}
}

//让Router具有处理请求的功能。
func (this *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	//每个请求的入口在这里
	//全局异常处理。
	defer this.GlobalPanicHandler(writer, request)

	path := request.URL.Path
	if strings.HasPrefix(path, "/api") {

		if handler, ok := this.routeMap[path]; ok {

			handler(writer, request)

		} else {
			//直接将请求扔给每个controller，看看他们能不能处理，如果都不能处理，那就算了。
			canHandle := false
			for _, controller := range this.context.ControllerMap {
				if handler, exist := controller.HandleRoutes(writer, request); exist {
					canHandle = true

					handler(writer, request)
					break
				}
			}

			if !canHandle {
				panic(fmt.Sprintf("没有找到能够处理%s的方法\n", path))
			}

		}

	} else {
		//当作静态资源处理。默认从当前文件下面的static文件夹中取东西。
		dir := GetHtmlPath()

		requestURI := request.RequestURI
		if requestURI == "" || request.RequestURI == "/" {
			requestURI = "index.html"
		}

		filePath := dir + requestURI
		exists, _ := PathExists(filePath)
		if !exists {
			filePath = dir + "/index.html"
			exists, _ = PathExists(filePath)
			if !exists {
				panic(fmt.Sprintf("404 not found:%s", filePath))
			}
		}

		writer.Header().Set("Content-Type", GetMimeType(GetExtension(filePath)))

		diskFile, err := os.Open(filePath)
		if err != nil {
			panic("cannot get file.")
		}
		defer diskFile.Close()
		_, err = io.Copy(writer, diskFile)
		if err != nil {
			panic("cannot get file.")
		}

	}

}
