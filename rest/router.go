package rest

import (
	"fmt"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"os"
	"strings"
	"tank/rest/config"
	"tank/rest/result"
	"tank/rest/tool"
	"time"
)

//用于处理所有前来的请求
type Router struct {
	installController *InstallController
	footprintService  *FootprintService
	userService       *UserService
	routeMap          map[string]func(writer http.ResponseWriter, request *http.Request)
	installRouteMap   map[string]func(writer http.ResponseWriter, request *http.Request)
}

//构造方法
func NewRouter() *Router {
	router := &Router{
		routeMap:        make(map[string]func(writer http.ResponseWriter, request *http.Request)),
		installRouteMap: make(map[string]func(writer http.ResponseWriter, request *http.Request)),
	}

	//installController.
	b := CONTEXT.GetBean(router.installController)
	if b, ok := b.(*InstallController); ok {
		router.installController = b
	}

	//装载userService.
	b = CONTEXT.GetBean(router.userService)
	if b, ok := b.(*UserService); ok {
		router.userService = b
	}

	//装载footprintService
	b = CONTEXT.GetBean(router.footprintService)
	if b, ok := b.(*FootprintService); ok {
		router.footprintService = b
	}

	//将Controller中的路由规则装载进来，InstallController中的除外
	for _, controller := range CONTEXT.ControllerMap {

		if controller == router.installController {
			routes := controller.RegisterRoutes()
			for k, v := range routes {
				router.installRouteMap[k] = v
			}
		} else {
			routes := controller.RegisterRoutes()
			for k, v := range routes {
				router.routeMap[k] = v
			}
		}

	}
	return router

}

//全局的异常捕获
func (this *Router) GlobalPanicHandler(writer http.ResponseWriter, request *http.Request, startTime time.Time) {
	if err := recover(); err != nil {

		tool.LOGGER.Error("错误: %v", err)

		var webResult *result.WebResult = nil
		if value, ok := err.(string); ok {
			//一个字符串，默认是请求错误。
			webResult = result.CustomWebResult(result.CODE_WRAPPER_BAD_REQUEST, value)
		} else if value, ok := err.(*result.WebResult); ok {
			//一个WebResult对象
			webResult = value
		} else if value, ok := err.(*result.CodeWrapper); ok {
			//一个WebResult对象
			webResult = result.ConstWebResult(value)
		} else if value, ok := err.(error); ok {
			//一个普通的错误对象
			webResult = result.CustomWebResult(result.CODE_WRAPPER_UNKNOWN, value.Error())
		} else {
			//其他不能识别的内容
			webResult = result.ConstWebResult(result.CODE_WRAPPER_UNKNOWN)
		}

		//修改http code码
		writer.WriteHeader(result.FetchHttpStatus(webResult.Code))

		//输出的是json格式 返回的内容申明是json，utf-8
		writer.Header().Set("Content-Type", "application/json;charset=UTF-8")

		//用json的方式输出返回值。
		b, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(webResult)

		//写到输出流中
		_, err := fmt.Fprintf(writer, string(b))
		if err != nil {
			fmt.Printf("输出结果时出错了\n")
		}

		//错误情况记录。
		go tool.SafeMethod(func() {
			this.footprintService.Trace(writer, request, time.Now().Sub(startTime), false)
		})
	}
}

//让Router具有处理请求的功能。
func (this *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	startTime := time.Now()

	//每个请求的入口在这里
	//全局异常处理。
	defer this.GlobalPanicHandler(writer, request, startTime)

	path := request.URL.Path
	if strings.HasPrefix(path, "/api") {

		//对于IE浏览器，会自动缓存，因此设置不缓存Header.
		writer.Header().Set("Pragma", "No-cache")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Expires", "0")


		if config.CONFIG.Installed {
			//已安装的模式

			//统一处理用户的身份信息。
			this.userService.bootstrap(writer, request)

			if handler, ok := this.routeMap[path]; ok {
				handler(writer, request)
			} else {
				//直接将请求扔给每个controller，看看他们能不能处理，如果都不能处理，那就抛出找不到的错误
				canHandle := false
				for _, controller := range CONTEXT.ControllerMap {
					if handler, exist := controller.HandleRoutes(writer, request); exist {
						canHandle = true
						handler(writer, request)
						break
					}
				}

				if !canHandle {
					panic(result.CustomWebResult(result.CODE_WRAPPER_NOT_FOUND, fmt.Sprintf("没有找到能够处理%s的方法", path)))
				}
			}

			//正常的访问记录会落到这里。
			go tool.SafeMethod(func() {
				this.footprintService.Trace(writer, request, time.Now().Sub(startTime), true)
			})

		} else {
			//未安装模式
			if handler, ok := this.installRouteMap[path]; ok {
				handler(writer, request)
			} else {
				panic(result.ConstWebResult(result.CODE_WRAPPER_NOT_INSTALLED))
			}
		}

	} else {
		//当作静态资源处理。默认从当前文件下面的static文件夹中取东西。
		dir := tool.GetHtmlPath()

		requestURI := request.RequestURI
		if requestURI == "" || request.RequestURI == "/" {
			requestURI = "index.html"
		}

		filePath := dir + requestURI
		exists, _ := tool.PathExists(filePath)
		if !exists {
			filePath = dir + "/index.html"
			exists, _ = tool.PathExists(filePath)
			if !exists {
				panic(fmt.Sprintf("404 not found:%s", filePath))
			}
		}

		writer.Header().Set("Content-Type", tool.GetMimeType(tool.GetExtension(filePath)))

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
