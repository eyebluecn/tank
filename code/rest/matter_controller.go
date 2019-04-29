package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
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
	bridgeDao         *BridgeDao
	imageCacheService *ImageCacheService
}

//初始化方法 start to develop v3.
func (this *MatterController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
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

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if b, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = b
	}

}

//注册自己的路由。
func (this *MatterController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
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

	//本地文件映射
	routeMap["/api/matter/mirror"] = this.Wrap(this.Mirror, USER_ROLE_USER)

	return routeMap
}

//查看某个文件的详情。
func (this *MatterController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("文件的uuid必填"))
	}

	matter := this.matterService.Detail(uuid)

	//验证当前之人是否有权限查看这么详细。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		if matter.UserUuid != user.Uuid {
			panic("没有权限查看该文件")
		}
	}

	return this.Success(matter)

}

//按照分页的方式获取某个文件夹下文件和子文件夹的列表，通常情况下只有一页。
func (this *MatterController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//如果是根目录，那么就传入root.
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")
	orderTimes := request.FormValue("orderTimes")

	puuid := request.FormValue("puuid")
	userUuid := request.FormValue("userUuid")
	name := request.FormValue("name")
	dir := request.FormValue("dir")
	orderDir := request.FormValue("orderDir")
	orderSize := request.FormValue("orderSize")
	orderName := request.FormValue("orderName")
	extensionsStr := request.FormValue("extensions")

	//使用分享提取码的形式授权。
	shareUuid := request.FormValue("shareUuid")
	shareCode := request.FormValue("shareCode")
	shareRootUuid := request.FormValue("shareRootUuid")
	if shareUuid != "" {

		if puuid == "" {
			panic(result.BadRequest("puuid必填！"))
		}
		dirMatter := this.matterDao.CheckByUuid(puuid)
		if !dirMatter.Dir {
			panic(result.BadRequest("puuid 对应的不是文件夹"))
		}

		share := this.shareDao.CheckByUuid(shareUuid)
		//如果是自己的分享，可以不要提取码
		user := this.findUser(writer, request)
		if user == nil {
			if share.Code != shareCode {
				panic(result.Unauthorized("提取码错误！"))
			}
		} else {
			if user.Uuid != share.UserUuid {
				if share.Code != shareCode {
					panic(result.Unauthorized("提取码错误！"))
				}
			}
		}

		//验证 shareRootMatter是否在被分享。
		shareRootMatter := this.matterDao.CheckByUuid(shareRootUuid)
		if !shareRootMatter.Dir {
			panic(result.BadRequest("只有文件夹可以浏览！"))
		}
		this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, shareRootMatter.Uuid)

		//保证 puuid对应的matter是shareRootMatter的子文件夹。
		child := strings.HasPrefix(dirMatter.Path, shareRootMatter.Path)
		if !child {
			panic(result.BadRequest("%s 不是 %s 的子文件夹！", puuid, shareRootUuid))
		}

	} else {
		//非分享模式要求必须登录
		user := this.checkUser(writer, request)
		if user.Role != USER_ROLE_ADMINISTRATOR {
			userUuid = user.Uuid
		}
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

	//筛选后缀名
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

//创建一个文件夹。
func (this *MatterController) CreateDirectory(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	puuid := request.FormValue("puuid")
	name := request.FormValue("name")
	userUuid := request.FormValue("userUuid")

	//管理员可以指定给某个用户创建文件夹。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}
	user = this.userDao.CheckByUuid(userUuid)

	//找到父级matter
	var dirMatter *Matter
	if puuid == MATTER_ROOT {
		dirMatter = NewRootMatter(user)
	} else {
		dirMatter = this.matterDao.CheckByUuid(puuid)
	}

	matter := this.matterService.AtomicCreateDirectory(dirMatter, name, user)
	return this.Success(matter)
}

//上传文件
func (this *MatterController) Upload(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	userUuid := request.FormValue("userUuid")
	puuid := request.FormValue("puuid")
	privacyStr := request.FormValue("privacy")
	file, handler, err := request.FormFile("file")
	this.PanicError(err)
	defer func() {
		err := file.Close()
		this.PanicError(err)
	}()

	user := this.checkUser(writer, request)
	//管理员可以传到指定用户的目录下。
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}
	user = this.userDao.CheckByUuid(userUuid)

	privacy := privacyStr == TRUE

	err = request.ParseMultipartForm(32 << 20)
	this.PanicError(err)

	//对于IE浏览器，filename可能包含了路径。
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

	matter := this.matterService.AtomicUpload(file, user, dirMatter, fileName, privacy)

	return this.Success(matter)
}

//从一个Url中去爬取资源
func (this *MatterController) Crawl(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	url := request.FormValue("url")
	destPath := request.FormValue("destPath")
	userUuid := request.FormValue("userUuid")
	filename := request.FormValue("filename")

	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	} else {
		if userUuid == "" {
			userUuid = user.Uuid
		}
	}

	user = this.userDao.CheckByUuid(userUuid)

	dirMatter := this.matterService.CreateDirectories(user, destPath)

	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		panic("资源url必填，并且应该以http://或者https://开头")
	}

	if filename == "" {
		panic("filename 必填")
	}

	matter := this.matterService.AtomicCrawl(url, filename, user, dirMatter, true)

	return this.Success(matter)
}

//删除一个文件
func (this *MatterController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("文件的uuid必填"))
	}

	matter := this.matterDao.CheckByUuid(uuid)

	//判断文件的所属人是否正确
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
		panic(result.Unauthorized("没有权限"))
	}

	this.matterService.AtomicDelete(matter)

	return this.Success("删除成功！")
}

//删除一系列文件。
func (this *MatterController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("文件的uuids必填"))
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		matter := this.matterDao.FindByUuid(uuid)

		//如果matter已经是nil了，直接跳过
		if matter == nil {
			this.logger.Warn("%s 对应的文件记录已经不存在了", uuid)
			continue
		}

		//判断文件的所属人是否正确
		user := this.checkUser(writer, request)
		if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
			panic(result.Unauthorized("没有权限"))
		}

		this.matterService.AtomicDelete(matter)

	}

	return this.Success("删除成功！")
}

//重命名一个文件或一个文件夹
func (this *MatterController) Rename(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	name := request.FormValue("name")

	user := this.checkUser(writer, request)

	//找出该文件或者文件夹
	matter := this.matterDao.CheckByUuid(uuid)

	if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
		panic(result.Unauthorized("没有权限"))
	}

	this.matterService.AtomicRename(matter, name, user)

	return this.Success(matter)
}

//改变一个文件的公私有属性
func (this *MatterController) ChangePrivacy(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuid := request.FormValue("uuid")
	privacyStr := request.FormValue("privacy")
	privacy := false
	if privacyStr == TRUE {
		privacy = true
	}
	//找出该文件或者文件夹
	matter := this.matterDao.CheckByUuid(uuid)

	if matter.Privacy == privacy {
		panic("公私有属性没有改变！")
	}

	//权限验证
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
		panic(result.Unauthorized("没有权限"))
	}

	matter.Privacy = privacy
	this.matterDao.Save(matter)

	return this.Success("设置成功")
}

//将一个文件夹或者文件移入到另一个文件夹下。
func (this *MatterController) Move(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	srcUuidsStr := request.FormValue("srcUuids")
	destUuid := request.FormValue("destUuid")
	userUuid := request.FormValue("userUuid")

	var srcUuids []string
	//验证参数。
	if srcUuidsStr == "" {
		panic(result.BadRequest("srcUuids参数必填"))
	} else {
		srcUuids = strings.Split(srcUuidsStr, ",")
	}

	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR || userUuid == "" {
		userUuid = user.Uuid
	}

	user = this.userDao.CheckByUuid(userUuid)

	//验证dest是否有问题
	var destMatter = this.matterDao.CheckWithRootByUuid(destUuid, user)
	if !destMatter.Dir {
		panic(result.BadRequest("目标不是文件夹"))
	}

	if user.Role != USER_ROLE_ADMINISTRATOR && destMatter.UserUuid != user.Uuid {
		panic(result.Unauthorized("没有权限"))
	}

	var srcMatters []*Matter
	//验证src是否有问题。
	for _, uuid := range srcUuids {
		//找出该文件或者文件夹
		srcMatter := this.matterDao.CheckByUuid(uuid)

		if srcMatter.Puuid == destMatter.Uuid {
			panic(result.BadRequest("没有进行移动，操作无效！"))
		}

		//判断同级文件夹中是否有同名的文件
		count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, destMatter.Uuid, srcMatter.Dir, srcMatter.Name)

		if count > 0 {
			panic(result.BadRequest("【" + srcMatter.Name + "】在目标文件夹已经存在了，操作失败。"))
		}

		//判断和目标文件夹是否是同一个主人。
		if srcMatter.UserUuid != destMatter.UserUuid {
			panic("文件和目标文件夹的拥有者不是同一人")
		}

		srcMatters = append(srcMatters, srcMatter)
	}

	this.matterService.AtomicMoveBatch(srcMatters, destMatter)

	return this.Success(nil)
}

//将本地文件映射到蓝眼云盘中去。
func (this *MatterController) Mirror(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	srcPath := request.FormValue("srcPath")
	destPath := request.FormValue("destPath")
	overwriteStr := request.FormValue("overwrite")

	if srcPath == "" {
		panic(result.BadRequest("srcPath必填"))
	}

	overwrite := false
	if overwriteStr == TRUE {
		overwrite = true
	}

	user := this.userDao.checkUser(writer, request)

	this.matterService.AtomicMirror(srcPath, destPath, overwrite, user)

	return this.Success(nil)

}
