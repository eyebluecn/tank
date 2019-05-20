package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
	"strings"
)

type MatterController struct {
	BaseController
	matterDao         *MatterDao
	matterService     *MatterService
	downloadTokenDao  *DownloadTokenDao
	imageCacheDao     *ImageCacheDao
	shareDao          *ShareDao
	shareService      *ShareService
	bridgeDao         *BridgeDao
	imageCacheService *ImageCacheService
}

func (this *MatterController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if b, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if b, ok := b.(*ShareService); ok {
		this.shareService = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if b, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = b
	}
}

func (this *MatterController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/matter/create/directory"] = this.Wrap(this.CreateDirectory, USER_ROLE_USER)
	routeMap["/api/matter/upload"] = this.Wrap(this.Upload, USER_ROLE_USER)
	routeMap["/api/matter/crawl"] = this.Wrap(this.Crawl, USER_ROLE_USER)
	routeMap["/api/matter/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/matter/delete/batch"] = this.Wrap(this.DeleteBatch, USER_ROLE_USER)
	routeMap["/api/matter/rename"] = this.Wrap(this.Rename, USER_ROLE_USER)
	routeMap["/api/matter/change/privacy"] = this.Wrap(this.ChangePrivacy, USER_ROLE_USER)
	routeMap["/api/matter/move"] = this.Wrap(this.Move, USER_ROLE_USER)
	routeMap["/api/matter/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/matter/page"] = this.Wrap(this.Page, USER_ROLE_GUEST)

	//mirror local files.
	routeMap["/api/matter/mirror"] = this.Wrap(this.Mirror, USER_ROLE_USER)
	routeMap["/api/matter/zip"] = this.Wrap(this.Zip, USER_ROLE_USER)

	return routeMap
}

func (this *MatterController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	matter := this.matterService.Detail(request, uuid)

	user := this.checkUser(request)
	if matter.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(matter)

}

func (this *MatterController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")
	orderTimes := request.FormValue("orderTimes")

	puuid := request.FormValue("puuid")
	name := request.FormValue("name")
	dir := request.FormValue("dir")
	orderDir := request.FormValue("orderDir")
	orderSize := request.FormValue("orderSize")
	orderName := request.FormValue("orderName")
	extensionsStr := request.FormValue("extensions")

	var userUuid string

	//auth by shareUuid.
	shareUuid := request.FormValue("shareUuid")
	shareCode := request.FormValue("shareCode")
	shareRootUuid := request.FormValue("shareRootUuid")
	if shareUuid != "" {

		if puuid == "" {
			panic(result.BadRequest("puuid cannot be null"))
		}

		dirMatter := this.matterDao.CheckByUuid(puuid)
		if !dirMatter.Dir {
			panic(result.BadRequest("puuid is not a directory"))
		}

		user := this.findUser(request)

		this.shareService.ValidateMatter(request, shareUuid, shareCode, user, shareRootUuid, dirMatter)
		userUuid = dirMatter.Uuid

	} else {
		//if cannot auth by share. Then login is required.
		user := this.checkUser(request)
		userUuid = user.Uuid

	}

	var page int
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	pageSize := 200
	if pageSizeStr != "" {
		tmp, err := strconv.Atoi(pageSizeStr)
		if err == nil {
			pageSize = tmp
		}
	}

	var extensions []string
	if extensionsStr != "" {
		extensions = strings.Split(extensionsStr, ",")
	}

	sortArray := []builder.OrderPair{
		{
			Key:   "dir",
			Value: orderDir,
		},
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
		{
			Key:   "update_time",
			Value: orderUpdateTime,
		},
		{
			Key:   "sort",
			Value: orderSort,
		},
		{
			Key:   "size",
			Value: orderSize,
		},
		{
			Key:   "name",
			Value: orderName,
		},
		{
			Key:   "times",
			Value: orderTimes,
		},
	}

	pager := this.matterDao.Page(page, pageSize, puuid, userUuid, name, dir, extensions, sortArray)

	return this.Success(pager)
}

func (this *MatterController) CreateDirectory(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	puuid := request.FormValue("puuid")
	name := request.FormValue("name")

	user := this.checkUser(request)

	var dirMatter = this.matterDao.CheckWithRootByUuid(puuid, user)

	matter := this.matterService.AtomicCreateDirectory(request, dirMatter, name, user)
	return this.Success(matter)
}

func (this *MatterController) Upload(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	puuid := request.FormValue("puuid")
	privacyStr := request.FormValue("privacy")
	file, handler, err := request.FormFile("file")
	this.PanicError(err)
	defer func() {
		err := file.Close()
		this.PanicError(err)
	}()

	user := this.checkUser(request)

	privacy := privacyStr == TRUE

	err = request.ParseMultipartForm(32 << 20)
	this.PanicError(err)

	//for IE browser. filename may contains filepath.
	fileName := handler.Filename
	pos := strings.LastIndex(fileName, "\\")
	if pos != -1 {
		fileName = fileName[pos+1:]
	}
	pos = strings.LastIndex(fileName, "/")
	if pos != -1 {
		fileName = fileName[pos+1:]
	}

	dirMatter := this.matterDao.CheckWithRootByUuid(puuid, user)

	//support upload simultaneously
	matter := this.matterService.Upload(request, file, user, dirMatter, fileName, privacy)

	return this.Success(matter)
}

//crawl a file by url.
func (this *MatterController) Crawl(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	url := request.FormValue("url")
	destPath := request.FormValue("destPath")
	filename := request.FormValue("filename")

	user := this.checkUser(request)

	dirMatter := this.matterService.CreateDirectories(request, user, destPath)

	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		panic(" url must start with  http:// or https://")
	}

	if filename == "" {
		panic("filename cannot be null")
	}

	matter := this.matterService.AtomicCrawl(request, url, filename, user, dirMatter, true)

	return this.Success(matter)
}

func (this *MatterController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	matter := this.matterDao.CheckByUuid(uuid)

	user := this.checkUser(request)
	if matter.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.matterService.AtomicDelete(request, matter, user)

	return this.Success("OK")
}

func (this *MatterController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("uuids cannot be null"))
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		matter := this.matterDao.FindByUuid(uuid)

		if matter == nil {
			this.logger.Warn("%s not exist anymore", uuid)
			continue
		}

		user := this.checkUser(request)
		if matter.UserUuid != user.Uuid {
			panic(result.UNAUTHORIZED)
		}

		this.matterService.AtomicDelete(request, matter, user)

	}

	return this.Success("OK")
}

func (this *MatterController) Rename(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	name := request.FormValue("name")

	user := this.checkUser(request)

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	this.matterService.AtomicRename(request, matter, name, user)

	return this.Success(matter)
}

func (this *MatterController) ChangePrivacy(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuid := request.FormValue("uuid")
	privacyStr := request.FormValue("privacy")
	privacy := false
	if privacyStr == TRUE {
		privacy = true
	}

	matter := this.matterDao.CheckByUuid(uuid)

	if matter.Privacy == privacy {
		panic(result.BadRequest("not changed. Invalid operation."))
	}

	user := this.checkUser(request)
	if matter.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	matter.Privacy = privacy
	this.matterDao.Save(matter)

	return this.Success("OK")
}

func (this *MatterController) Move(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	srcUuidsStr := request.FormValue("srcUuids")
	destUuid := request.FormValue("destUuid")

	var srcUuids []string
	if srcUuidsStr == "" {
		panic(result.BadRequest("srcUuids cannot be null"))
	} else {
		srcUuids = strings.Split(srcUuidsStr, ",")
	}

	user := this.checkUser(request)

	var destMatter = this.matterDao.CheckWithRootByUuid(destUuid, user)
	if !destMatter.Dir {
		panic(result.BadRequest("destination is not a directory"))
	}

	if destMatter.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	var srcMatters []*Matter
	for _, uuid := range srcUuids {
		srcMatter := this.matterDao.CheckByUuid(uuid)

		if srcMatter.Puuid == destMatter.Uuid {
			panic(result.BadRequest("no move, invalid operation"))
		}

		//check whether there are files with the same name.
		count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, destMatter.Uuid, srcMatter.Dir, srcMatter.Name)

		if count > 0 {
			panic(result.BadRequestI18n(request, i18n.MatterExist, srcMatter.Name))
		}

		if srcMatter.UserUuid != destMatter.UserUuid {
			panic("owner not the same")
		}

		srcMatters = append(srcMatters, srcMatter)
	}

	this.matterService.AtomicMoveBatch(request, srcMatters, destMatter, user)

	return this.Success(nil)
}

//mirror local files to EyeblueTank
func (this *MatterController) Mirror(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	srcPath := request.FormValue("srcPath")
	destPath := request.FormValue("destPath")
	overwriteStr := request.FormValue("overwrite")

	if srcPath == "" {
		panic(result.BadRequest("srcPath cannot be null"))
	}

	overwrite := false
	if overwriteStr == TRUE {
		overwrite = true
	}

	user := this.userDao.checkUser(request)

	this.matterService.AtomicMirror(request, srcPath, destPath, overwrite, user)

	return this.Success(nil)

}

//download zip.
func (this *MatterController) Zip(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("uuids cannot be null"))
	}

	uuidArray := strings.Split(uuids, ",")

	matters := this.matterDao.FindByUuids(uuidArray, nil)

	if matters == nil || len(matters) == 0 {
		panic(result.BadRequest("matters cannot be nil."))
	}
	user := this.checkUser(request)
	puuid := matters[0].Puuid

	for _, m := range matters {
		if m.UserUuid != user.Uuid {
			panic(result.UNAUTHORIZED)
		} else if m.Puuid != puuid {
			panic(result.BadRequest("puuid not same"))
		}
	}

	this.matterService.DownloadZip(writer, request, matters)

	return nil
}
