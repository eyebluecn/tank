package rest

import (
	"bytes"
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

/**
 *
 * WebDav document
 * https://tools.ietf.org/html/rfc4918
 * http://www.webdav.org/specs/rfc4918.html
 *
 */

type DavController struct {
	BaseController
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	matterDao         *MatterDao
	matterService     *MatterService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
	davService        *DavService
}

func (this *DavController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if c, ok := b.(*MatterDao); ok {
		this.matterDao = c
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if c, ok := b.(*MatterService); ok {
		this.matterService = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}

	b = core.CONTEXT.GetBean(this.davService)
	if c, ok := b.(*DavService); ok {
		this.davService = c
	}
}

//Auth user by BasicAuth
func (this *DavController) CheckCurrentUser(writer http.ResponseWriter, request *http.Request) *User {

	username, password, ok := request.BasicAuth()
	if !ok {
		// require the basic auth.
		writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		panic(result.ConstWebResult(result.LOGIN))
	}

	user := this.userDao.FindByUsername(username)
	if user == nil {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	} else {
		if !util.MatchBcrypt(password, user.Password) {
			panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
		}
	}

	return user
}

func (this *DavController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	return routeMap
}

//handle some special routes, eg. params in the url.
func (this *DavController) HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool) {

	path := request.URL.Path

	//match /api/dav{subPath}
	pattern := fmt.Sprintf(`^%s(.*)$`, WEBDAV_PREFIX)
	reg := regexp.MustCompile(pattern)
	strs := reg.FindStringSubmatch(path)
	if len(strs) == 2 {
		var f = func(writer http.ResponseWriter, request *http.Request) {
			subPath := strs[1]
			//guarantee subPath not end with /
			subPath = strings.TrimSuffix(subPath, "/")
			this.Index(writer, request, subPath)
		}
		return f, true
	}

	return nil, false
}

func (this *DavController) debug(writer http.ResponseWriter, request *http.Request, subPath string) {

	//Print the Request info.
	fmt.Printf("\n------  %s  --  %s  ------\n", request.URL, subPath)

	fmt.Printf("\n------Method：------\n")
	fmt.Println(request.Method)

	fmt.Printf("\n------Header：------\n")
	for key, value := range request.Header {
		fmt.Printf("%s = %s\n", key, value)
	}

	fmt.Printf("\n------Params：------\n")
	for key, value := range request.Form {
		fmt.Printf("%s = %s\n", key, value)
	}

	fmt.Printf("\n------Body：------\n")
	//ioutil.ReadAll cannot read again. when read again, there is nothing.

	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println("读取body时出错" + err.Error())
	}
	fmt.Println(string(bodyBytes))

	//close and resign
	err = request.Body.Close()
	if err != nil {
		panic(err)
	}
	request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	fmt.Println("------------------")

}

func (this *DavController) Index(writer http.ResponseWriter, request *http.Request, subPath string) {

	//this.debug(writer, request, subPath)

	user := this.CheckCurrentUser(writer, request)

	this.davService.HandleDav(writer, request, user, subPath)

}
