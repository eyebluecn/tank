package rest

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AlienController struct {
	BaseController
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	matterDao         *MatterDao
	matterService     *MatterService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
	alienService      *AlienService
}

//初始化方法
func (this *AlienController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean.
	b := CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = CONTEXT.GetBean(this.matterDao)
	if c, ok := b.(*MatterDao); ok {
		this.matterDao = c
	}

	b = CONTEXT.GetBean(this.matterService)
	if c, ok := b.(*MatterService); ok {
		this.matterService = c
	}

	b = CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}

	b = CONTEXT.GetBean(this.alienService)
	if c, ok := b.(*AlienService); ok {
		this.alienService = c
	}

}

//注册自己的路由。
func (this *AlienController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/alien/fetch/upload/token"] = this.Wrap(this.FetchUploadToken, USER_ROLE_GUEST)
	routeMap["/api/alien/fetch/download/token"] = this.Wrap(this.FetchDownloadToken, USER_ROLE_GUEST)
	routeMap["/api/alien/confirm"] = this.Wrap(this.Confirm, USER_ROLE_GUEST)
	routeMap["/api/alien/upload"] = this.Wrap(this.Upload, USER_ROLE_GUEST)
	routeMap["/api/alien/crawl/token"] = this.Wrap(this.CrawlToken, USER_ROLE_GUEST)
	routeMap["/api/alien/crawl/direct"] = this.Wrap(this.CrawlDirect, USER_ROLE_GUEST)

	return routeMap
}

//处理一些特殊的接口，比如参数包含在路径中,一般情况下，controller不将参数放在url路径中
func (this *AlienController) HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool) {

	path := request.URL.Path

	//匹配 /api/alien/preview/{uuid}/{filename} (响应头不包含 content-disposition)
	reg := regexp.MustCompile(`^/api/alien/preview/([^/]+)/([^/]+)$`)
	strs := reg.FindStringSubmatch(path)
	if len(strs) == 3 {
		var f = func(writer http.ResponseWriter, request *http.Request) {
			this.Preview(writer, request, strs[1], strs[2])
		}
		return f, true
	}

	//匹配 /api/alien/download/{uuid}/{filename} (响应头包含 content-disposition)
	reg = regexp.MustCompile(`^/api/alien/download/([^/]+)/([^/]+)$`)
	strs = reg.FindStringSubmatch(path)
	if len(strs) == 3 {
		var f = func(writer http.ResponseWriter, request *http.Request) {
			this.Download(writer, request, strs[1], strs[2])
		}
		return f, true
	}

	return nil, false
}

//直接从cookie中获取用户信息，或者使用邮箱和密码获取用户
func (this *AlienController) CheckRequestUser(writer http.ResponseWriter, request *http.Request) *User {

	//根据用户登录信息取
	user := this.findUser(writer, request)
	if user != nil {
		return user;
	}

	email := request.FormValue("email")
	if email == "" {
		panic("邮箱必填啦")
	}

	password := request.FormValue("password")
	if password == "" {
		panic("密码必填")
	}

	//验证用户身份合法性。
	user = this.userDao.FindByEmail(email)
	if user == nil {
		panic(`邮箱或密码错误`)
	} else {
		if !MatchBcrypt(password, user.Password) {
			panic(`邮箱或密码错误`)
		}
	}
	return user
}

//系统中的用户x要获取一个UploadToken，用于提供给x信任的用户上传文件。
func (this *AlienController) FetchUploadToken(writer http.ResponseWriter, request *http.Request) *WebResult {

	//文件名。
	filename := request.FormValue("filename")
	if filename == "" {
		panic("文件名必填")
	} else if m, _ := regexp.MatchString(`[<>|*?/\\]`, filename); m {
		panic(fmt.Sprintf(`【%s】不符合要求，文件名中不能包含以下特殊符号：< > | * ? / \`, filename))
	}

	//什么时间后过期，默认24h
	expireStr := request.FormValue("expire")
	expire := 24 * 60 * 60
	if expireStr != "" {
		var err error
		expire, err = strconv.Atoi(expireStr)
		if err != nil {
			panic(`过期时间不符合规范`)
		}
		if expire < 1 {
			panic(`过期时间不符合规范`)
		}

	}

	//文件公有或私有
	privacyStr := request.FormValue("privacy")
	var privacy bool
	if privacyStr == "" {
		panic(`文件公有性必填`)
	} else {
		if privacyStr == "true" {
			privacy = true
		} else if privacyStr == "false" {
			privacy = false
		} else {
			panic(`文件公有性不符合规范`)
		}
	}

	//文件大小
	sizeStr := request.FormValue("size")
	var size int64
	if sizeStr == "" {
		panic(`文件大小必填`)
	} else {

		var err error
		size, err = strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			panic(`文件大小不符合规范`)
		}
		if size < 1 {
			panic(`文件大小不符合规范`)
		}
	}

	//文件夹路径，以 / 开头。
	dir := request.FormValue("dir")

	user := this.CheckRequestUser(writer, request)
	dirUuid := this.matterService.GetDirUuid(user.Uuid, dir)

	mm, _ := time.ParseDuration(fmt.Sprintf("%ds", expire))
	uploadToken := &UploadToken{
		UserUuid:   user.Uuid,
		FolderUuid: dirUuid,
		MatterUuid: "",
		ExpireTime: time.Now().Add(mm),
		Filename:   filename,
		Privacy:    privacy,
		Size:       size,
		Ip:         GetIpAddress(request),
	}

	uploadToken = this.uploadTokenDao.Create(uploadToken)

	return this.Success(uploadToken)

}

//系统中的用户x 拿着某个文件的uuid来确认是否其信任的用户已经上传好了。
func (this *AlienController) Confirm(writer http.ResponseWriter, request *http.Request) *WebResult {

	matterUuid := request.FormValue("matterUuid")
	if matterUuid == "" {
		panic("matterUuid必填")
	}

	user := this.CheckRequestUser(writer, request)

	matter := this.matterDao.CheckByUuid(matterUuid)
	if matter.UserUuid != user.Uuid {
		panic("文件不属于你")
	}

	return this.Success(matter)
}

//系统中的用户x 信任的用户上传文件。这个接口需要支持跨域。
func (this *AlienController) Upload(writer http.ResponseWriter, request *http.Request) *WebResult {
	//允许跨域请求。
	this.allowCORS(writer)
	if request.Method == "OPTIONS" {
		return this.Success("OK")
	}

	uploadTokenUuid := request.FormValue("uploadTokenUuid")
	if uploadTokenUuid == "" {
		panic("uploadTokenUuid必填")
	}

	uploadToken := this.uploadTokenDao.FindByUuid(uploadTokenUuid)
	if uploadToken == nil {
		panic("uploadTokenUuid无效")
	}

	if uploadToken.ExpireTime.Before(time.Now()) {
		panic("uploadToken已失效")
	}

	user := this.userDao.CheckByUuid(uploadToken.UserUuid)

	request.ParseMultipartForm(32 << 20)
	file, handler, err := request.FormFile("file")
	this.PanicError(err)
	defer file.Close()

	if handler.Filename != uploadToken.Filename {
		panic("文件名称不正确")
	}

	if handler.Size != uploadToken.Size {
		panic("文件大小不正确")
	}

	matter := this.matterService.Upload(file, user, uploadToken.FolderUuid, uploadToken.Filename, uploadToken.Privacy, true)

	//更新这个uploadToken的信息.
	uploadToken.ExpireTime = time.Now()
	this.uploadTokenDao.Save(uploadToken)

	return this.Success(matter)
}

//给一个指定的url，从该url中去拉取文件回来。此处采用uploadToken的模式。
func (this *AlienController) CrawlToken(writer http.ResponseWriter, request *http.Request) *WebResult {
	//允许跨域请求。
	this.allowCORS(writer)
	if request.Method == "OPTIONS" {
		return this.Success("OK")
	}

	uploadTokenUuid := request.FormValue("uploadTokenUuid")
	if uploadTokenUuid == "" {
		panic("uploadTokenUuid必填")
	}

	uploadToken := this.uploadTokenDao.FindByUuid(uploadTokenUuid)
	if uploadToken == nil {
		panic("uploadTokenUuid无效")
	}

	if uploadToken.ExpireTime.Before(time.Now()) {
		panic("uploadToken已失效")
	}

	user := this.userDao.CheckByUuid(uploadToken.UserUuid)

	url := request.FormValue("url")
	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		panic("资源url必填，并且应该以http://或者https://开头")
	}

	matter := this.matterService.Crawl(url, uploadToken.Filename, user, uploadToken.FolderUuid, uploadToken.Privacy)

	//更新这个uploadToken的信息.
	uploadToken.ExpireTime = time.Now()
	this.uploadTokenDao.Save(uploadToken)

	return this.Success(matter)
}

//通过一个url直接上传，无需借助uploadToken.
func (this *AlienController) CrawlDirect(writer http.ResponseWriter, request *http.Request) *WebResult {

	//文件名。
	filename := request.FormValue("filename")
	if filename == "" {
		panic("文件名必填")
	} else if m, _ := regexp.MatchString(`[<>|*?/\\]`, filename); m {
		panic(fmt.Sprintf(`【%s】不符合要求，文件名中不能包含以下特殊符号：< > | * ? / \`, filename))
	}

	url := request.FormValue("url")
	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		panic("资源url必填，并且应该以http://或者https://开头")
	}

	//文件公有或私有
	privacyStr := request.FormValue("privacy")
	var privacy bool
	if privacyStr == "" {
		panic(`文件公有性必填`)
	} else {
		if privacyStr == "true" {
			privacy = true
		} else if privacyStr == "false" {
			privacy = false
		} else {
			panic(`文件公有性不符合规范`)
		}
	}

	//文件夹路径，以 / 开头。
	dir := request.FormValue("dir")
	user := this.CheckRequestUser(writer, request)
	dirUuid := this.matterService.GetDirUuid(user.Uuid, dir)

	matter := this.matterService.Crawl(url, filename, user, dirUuid, privacy)

	return this.Success(matter)
}

//系统中的用户x要获取一个DownloadToken，用于提供给x信任的用户下载文件。
func (this *AlienController) FetchDownloadToken(writer http.ResponseWriter, request *http.Request) *WebResult {

	matterUuid := request.FormValue("matterUuid")
	if matterUuid == "" {
		panic("matterUuid必填")
	}

	user := this.CheckRequestUser(writer, request)

	matter := this.matterDao.CheckByUuid(matterUuid)
	if matter.UserUuid != user.Uuid {
		panic("文件不属于你")
	}
	if matter.Dir {
		panic("不支持下载文件夹")
	}

	//什么时间后过期，默认24h
	expireStr := request.FormValue("expire")
	expire := 24 * 60 * 60
	if expireStr != "" {
		var err error
		expire, err = strconv.Atoi(expireStr)
		if err != nil {
			panic(`过期时间不符合规范`)
		}
		if expire < 1 {
			panic(`过期时间不符合规范`)
		}

	}

	mm, _ := time.ParseDuration(fmt.Sprintf("%ds", expire))
	downloadToken := &DownloadToken{
		UserUuid:   user.Uuid,
		MatterUuid: matterUuid,
		ExpireTime: time.Now().Add(mm),
		Ip:         GetIpAddress(request),
	}

	downloadToken = this.downloadTokenDao.Create(downloadToken)

	return this.Success(downloadToken)

}

//预览一个文件。既可以使用登录的方式，也可以使用授权的方式
func (this *AlienController) Preview(writer http.ResponseWriter, request *http.Request, uuid string, filename string) {

	this.alienService.PreviewOrDownload(writer, request, uuid, filename,  false)
}

//下载一个文件。既可以使用登录的方式，也可以使用授权的方式
func (this *AlienController) Download(writer http.ResponseWriter, request *http.Request, uuid string, filename string) {

	this.alienService.PreviewOrDownload(writer, request, uuid, filename,  true)
}
