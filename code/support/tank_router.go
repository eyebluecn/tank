package support

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/rest"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type TankRouter struct {
	installController *rest.InstallController
	footprintService  *rest.FootprintService
	userService       *rest.UserService
	routeMap          map[string]func(writer http.ResponseWriter, request *http.Request)
	installRouteMap   map[string]func(writer http.ResponseWriter, request *http.Request)
}

func NewRouter() *TankRouter {
	router := &TankRouter{
		routeMap:        make(map[string]func(writer http.ResponseWriter, request *http.Request)),
		installRouteMap: make(map[string]func(writer http.ResponseWriter, request *http.Request)),
	}

	//installController.
	b := core.CONTEXT.GetBean(router.installController)
	if b, ok := b.(*rest.InstallController); ok {
		router.installController = b
	}

	//load userService
	b = core.CONTEXT.GetBean(router.userService)
	if b, ok := b.(*rest.UserService); ok {
		router.userService = b
	}

	//load footprintService
	b = core.CONTEXT.GetBean(router.footprintService)
	if b, ok := b.(*rest.FootprintService); ok {
		router.footprintService = b
	}

	//load Controllers except InstallController
	for _, controller := range core.CONTEXT.GetControllerMap() {

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

//catch global panic.
func (this *TankRouter) GlobalPanicHandler(writer http.ResponseWriter, request *http.Request, startTime time.Time) {
	if err := recover(); err != nil {

		//get panic file and line number.
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 0
		}

		//unkown panic
		if strings.HasSuffix(file, "runtime/panic.go") {
			_, file, line, ok = runtime.Caller(4)
			if !ok {
				file = "???"
				line = 0
			}
		}
		//async panic
		if strings.HasSuffix(file, "core/handler.go") {
			_, file, line, ok = runtime.Caller(4)
			if !ok {
				file = "???"
				line = 0
			}
		}

		core.LOGGER.Error("panic on %s:%d %v", util.GetFilenameOfPath(file), line, err)

		var webResult *result.WebResult = nil
		if value, ok := err.(string); ok {
			//string, default as BadRequest.
			webResult = result.CustomWebResult(result.BAD_REQUEST, value)
		} else if value, ok := err.(*result.WebResult); ok {
			//*result.WebResult
			webResult = value
		} else if value, ok := err.(*result.CodeWrapper); ok {
			//*result.CodeWrapper
			webResult = result.ConstWebResult(value)
		} else if value, ok := err.(error); ok {
			//normal error
			webResult = result.CustomWebResult(result.UNKNOWN, value.Error())
		} else {
			//other error
			webResult = result.ConstWebResult(result.UNKNOWN)
		}

		//change the http status.
		writer.WriteHeader(result.FetchHttpStatus(webResult.Code))

		//if json, set the Content-Type to json.
		writer.Header().Set("Content-Type", "application/json;charset=UTF-8")

		//write the response.
		b, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(webResult)

		//write to writer.
		_, err := fmt.Fprintf(writer, string(b))
		if err != nil {
			fmt.Printf("occur error while write response %s\r\n", err.Error())
		}

		//log error.
		go core.RunWithRecovery(func() {
			this.footprintService.Trace(request, time.Now().Sub(startTime), false)
		})
	}
}

func (this *TankRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	startTime := time.Now()

	//global panic handler
	defer this.GlobalPanicHandler(writer, request, startTime)

	path := request.URL.Path
	if strings.HasPrefix(path, "/api") {

		//IE browser will cache automatically. disable the cache.
		util.DisableCache(writer)

		if core.CONFIG.Installed() {

			//if installed.

			//handler user's auth info.
			this.userService.PreHandle(writer, request)

			if handler, ok := this.routeMap[path]; ok {
				handler(writer, request)
			} else {

				//dispatch the request to controller's handler.
				canHandle := false
				for _, controller := range core.CONTEXT.GetControllerMap() {
					if handler, exist := controller.HandleRoutes(writer, request); exist {
						canHandle = true
						handler(writer, request)
						break
					}
				}

				if !canHandle {
					panic(result.CustomWebResult(result.NOT_FOUND, fmt.Sprintf("cannot handle %s", path)))
				}
			}

			//log the request
			go core.RunWithRecovery(func() {
				this.footprintService.Trace(request, time.Now().Sub(startTime), true)
			})

		} else {
			//if not installed. try to install.
			if handler, ok := this.installRouteMap[path]; ok {
				handler(writer, request)
			} else {
				panic(result.ConstWebResult(result.NOT_INSTALLED))
			}
		}

	} else {

		//static file.
		dir := util.GetHtmlPath()

		if path == "" || path == "/" {
			path = "index.html"
		}

		filePath := dir + path
		exists := util.PathExists(filePath)
		if !exists {
			filePath = dir + "/index.html"
			exists = util.PathExists(filePath)
			if !exists {
				panic(fmt.Sprintf("404 not found:%s", filePath))
			}
		}

		writer.Header().Set("Content-Type", util.GetMimeType(util.GetExtension(filePath)))

		diskFile, err := os.Open(filePath)
		if err != nil {
			panic("cannot get file.")
		}
		defer func() {
			err := diskFile.Close()
			if err != nil {
				panic(err)
			}
		}()
		_, err = io.Copy(writer, diskFile)
		if err != nil {
			panic("cannot get file.")
		}

	}

}
