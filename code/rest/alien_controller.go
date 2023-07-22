package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type AlienController struct {
	BaseController
	uploadTokenDao     *UploadTokenDao
	downloadTokenDao   *DownloadTokenDao
	matterDao          *MatterDao
	spaceDao           *SpaceDao
	matterService      *MatterService
	imageCacheDao      *ImageCacheDao
	imageCacheService  *ImageCacheService
	alienService       *AlienService
	shareService       *ShareService
	spaceMemberService *SpaceMemberService
}

func (this *AlienController) Init() {
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

	b = core.CONTEXT.GetBean(this.spaceDao)
	if c, ok := b.(*SpaceDao); ok {
		this.spaceDao = c
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

	b = core.CONTEXT.GetBean(this.alienService)
	if c, ok := b.(*AlienService); ok {
		this.alienService = c
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if c, ok := b.(*ShareService); ok {
		this.shareService = c
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*SpaceMemberService); ok {
		this.spaceMemberService = b
	}
}

func (this *AlienController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/alien/fetch/upload/token"] = this.Wrap(this.FetchUploadToken, USER_ROLE_USER)
	routeMap["/api/alien/fetch/download/token"] = this.Wrap(this.FetchDownloadToken, USER_ROLE_USER)
	routeMap["/api/alien/confirm"] = this.Wrap(this.Confirm, USER_ROLE_USER)
	routeMap["/api/alien/upload"] = this.Wrap(this.Upload, USER_ROLE_GUEST)
	routeMap["/api/alien/crawl/token"] = this.Wrap(this.CrawlToken, USER_ROLE_GUEST)
	routeMap["/api/alien/crawl/direct"] = this.Wrap(this.CrawlDirect, USER_ROLE_USER)

	return routeMap
}

// handle some special routes, eg. params in the url.
func (this *AlienController) HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool) {

	path := request.URL.Path

	//match /api/alien/preview/{uuid}/{filename} (response header not contain content-disposition)
	reg := regexp.MustCompile(`^/api/alien/preview/([^/]+)/([^/]+)$`)
	strs := reg.FindStringSubmatch(path)
	if len(strs) == 3 {
		var f = func(writer http.ResponseWriter, request *http.Request) {
			this.Preview(writer, request, strs[1], strs[2])
		}
		return f, true
	}

	//match /api/alien/download/{uuid}/{filename} (response header contain content-disposition)
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

// fetch a upload token for guest. Guest can upload file with this token.
func (this *AlienController) FetchUploadToken(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	filename := request.FormValue("filename")
	expireTimeStr := request.FormValue("expireTime")
	privacyStr := request.FormValue("privacy")
	sizeStr := request.FormValue("size")
	//store dir path
	dirPath := request.FormValue("dirPath")

	filename = CheckMatterName(request, filename)

	var expireTime time.Time
	if expireTimeStr == "" {
		panic(result.BadRequest("time format error"))
	} else {
		expireTime = util.ConvertDateTimeStringToTime(expireTimeStr)
	}
	if expireTime.Before(time.Now()) {
		panic(result.BadRequest("expire time cannot before now"))
	}

	var privacy = false
	if privacyStr == TRUE {
		privacy = true
	}

	var size int64
	if sizeStr == "" {
		panic(result.BadRequest("file size cannot be null"))
	} else {

		var err error
		size, err = strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			panic(result.BadRequest("file size error"))
		}
		if size < 1 {
			panic(result.BadRequest("file size error"))
		}
	}

	user := this.checkUser(request)
	space := this.spaceDao.CheckByUuid(user.SpaceUuid)
	dirMatter := this.matterService.CreateDirectories(request, user, space, dirPath)

	uploadToken := &UploadToken{
		UserUuid:   user.Uuid,
		FolderUuid: dirMatter.Uuid,
		MatterUuid: "",
		ExpireTime: expireTime,
		Filename:   filename,
		Privacy:    privacy,
		Size:       size,
		Ip:         util.GetIpAddress(request),
	}

	uploadToken = this.uploadTokenDao.Create(uploadToken)

	return this.Success(uploadToken)

}

// user confirm a file whether uploaded successfully.
func (this *AlienController) Confirm(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	matterUuid := request.FormValue("matterUuid")
	if matterUuid == "" {
		panic(result.BadRequest("matterUuid  cannot be null"))
	}

	user := this.checkUser(request)

	matter := this.matterDao.CheckByUuid(matterUuid)
	if matter.UserUuid != user.Uuid {
		panic(result.BadRequest("matter not belong to you"))
	}

	return this.Success(matter)
}

// a guest upload a file with a upload token.
func (this *AlienController) Upload(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	//allow cors.
	this.allowCORS(writer)
	if request.Method == "OPTIONS" {
		//nil means empty response body.
		return nil
	}

	uploadTokenUuid := request.FormValue("uploadTokenUuid")
	file, handler, err := request.FormFile("file")
	this.PanicError(err)
	defer func() {
		e := file.Close()
		this.PanicError(e)
	}()

	if uploadTokenUuid == "" {
		panic(result.BadRequest("uploadTokenUuid cannot be null"))
	}

	uploadToken := this.uploadTokenDao.CheckByUuid(uploadTokenUuid)

	if uploadToken.ExpireTime.Before(time.Now()) {
		panic(result.BadRequest("uploadToken has expired"))
	}

	user := this.userDao.CheckByUuid(uploadToken.UserUuid)
	space := this.spaceDao.CheckByUuid(user.SpaceUuid)

	err = request.ParseMultipartForm(32 << 20)
	this.PanicError(err)

	if handler.Filename != uploadToken.Filename {
		panic(result.BadRequest("filename doesn't the one in uploadToken"))
	}

	if handler.Size != uploadToken.Size {
		panic(result.BadRequest("file size doesn't the one in uploadToken"))
	}

	dirMatter := this.matterDao.CheckWithRootByUuid(uploadToken.FolderUuid, space)

	matter := this.matterService.Upload(request, file, handler, user, space, dirMatter, uploadToken.Filename, uploadToken.Privacy)

	//expire the upload token.
	uploadToken.ExpireTime = time.Now()
	this.uploadTokenDao.Save(uploadToken)

	return this.Success(matter)
}

// crawl a url with uploadToken. guest can visit this method.
func (this *AlienController) CrawlToken(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//allow cors.
	this.allowCORS(writer)
	if request.Method == "OPTIONS" {
		//nil means empty response body.
		return nil
	}

	uploadTokenUuid := request.FormValue("uploadTokenUuid")
	url := request.FormValue("url")

	if uploadTokenUuid == "" {
		panic(result.BadRequest("uploadTokenUuid cannot be null"))
	}

	uploadToken := this.uploadTokenDao.CheckByUuid(uploadTokenUuid)

	if uploadToken.ExpireTime.Before(time.Now()) {
		panic(result.BadRequest("uploadToken has expired"))
	}

	user := this.userDao.CheckByUuid(uploadToken.UserUuid)
	space := this.spaceDao.CheckByUuid(user.SpaceUuid)

	dirMatter := this.matterDao.CheckWithRootByUuid(uploadToken.FolderUuid, space)

	matter := this.matterService.AtomicCrawl(request, url, uploadToken.Filename, user, space, dirMatter, uploadToken.Privacy)

	//expire the upload token.
	uploadToken.ExpireTime = time.Now()
	this.uploadTokenDao.Save(uploadToken)

	return this.Success(matter)
}

// crawl a url directly. only user can visit this method.
func (this *AlienController) CrawlDirect(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	filename := request.FormValue("filename")
	privacyStr := request.FormValue("privacy")
	dirPath := request.FormValue("dirPath")
	url := request.FormValue("url")

	filename = CheckMatterName(request, filename)

	var privacy bool
	if privacyStr == TRUE {
		privacy = true
	}

	user := this.checkUser(request)
	space := this.spaceDao.CheckByUuid(user.SpaceUuid)
	dirMatter := this.matterService.CreateDirectories(request, user, space, dirPath)

	matter := this.matterService.AtomicCrawl(request, url, filename, user, space, dirMatter, privacy)

	return this.Success(matter)
}

// fetch a download token for guest. Guest can download file with this token.
func (this *AlienController) FetchDownloadToken(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	matterUuid := request.FormValue("matterUuid")
	expireTimeStr := request.FormValue("expireTime")

	if matterUuid == "" {
		panic(result.BadRequest("matterUuid cannot be null."))
	}

	user := this.checkUser(request)

	matter := this.matterDao.CheckByUuid(matterUuid)
	canRead := this.spaceMemberService.canRead(user, matter.SpaceUuid)
	if !canRead {
		panic(result.BadRequest("no auth to visit this file."))
	}

	var expireTime time.Time
	if expireTimeStr == "" {
		panic(result.BadRequest("time format error"))
	} else {
		expireTime = util.ConvertDateTimeStringToTime(expireTimeStr)
	}
	if expireTime.Before(time.Now()) {
		panic(result.BadRequest("expire time cannot before now"))
	}

	downloadToken := &DownloadToken{
		UserUuid:   user.Uuid,
		MatterUuid: matterUuid,
		ExpireTime: expireTime,
		Ip:         util.GetIpAddress(request),
	}

	downloadToken = this.downloadTokenDao.Create(downloadToken)

	return this.Success(downloadToken)

}

// preview a file.
func (this *AlienController) Preview(writer http.ResponseWriter, request *http.Request, uuid string, filename string) {
	matter := this.alienService.ValidMatter(writer, request, uuid, filename)
	this.alienService.PreviewOrDownload(writer, request, matter, false)
}

// download a file.
func (this *AlienController) Download(writer http.ResponseWriter, request *http.Request, uuid string, filename string) {
	matter := this.alienService.ValidMatter(writer, request, uuid, filename)
	this.alienService.PreviewOrDownload(writer, request, matter, true)
}
