package rest

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"fmt"
)

type MatterController struct {
	BaseController
	matterDao        *MatterDao
	matterService    *MatterService
	downloadTokenDao *DownloadTokenDao
}

//初始化方法
func (this *MatterController) Init(context *Context) {
	this.BaseController.Init(context)

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := context.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = context.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = context.GetBean(this.downloadTokenDao)
	if b, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = b
	}

}

//注册自己的路由。
func (this *MatterController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/matter/create/directory"] = this.Wrap(this.CreateDirectory, USER_ROLE_USER)
	routeMap["/api/matter/upload"] = this.Wrap(this.Upload, USER_ROLE_USER)
	routeMap["/api/matter/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/matter/delete/batch"] = this.Wrap(this.DeleteBatch, USER_ROLE_USER)
	routeMap["/api/matter/rename"] = this.Wrap(this.Rename, USER_ROLE_USER)
	routeMap["/api/matter/change/privacy"] = this.Wrap(this.ChangePrivacy, USER_ROLE_USER)
	routeMap["/api/matter/move"] = this.Wrap(this.Move, USER_ROLE_USER)
	routeMap["/api/matter/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/matter/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

//查看某个文件的详情。
func (this *MatterController) Detail(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		return this.Error("文件的uuid必填")
	}

	matter := this.matterDao.FindByUuid(uuid)

	//组装file的内容，展示其父组件。
	puuid := matter.Puuid
	tmpMatter := matter
	for puuid != "root" {
		pFile := this.matterDao.FindByUuid(puuid)

		tmpMatter.Parent = pFile
		tmpMatter = pFile
		puuid = pFile.Puuid

	}

	return this.Success(matter)

}

//创建一个文件夹。
func (this *MatterController) CreateDirectory(writer http.ResponseWriter, request *http.Request) *WebResult {

	puuid := request.FormValue("puuid")

	name := request.FormValue("name")
	//验证参数。
	if name == "" {
		return this.Error("name参数必填")
	}
	if m, _ := regexp.MatchString(`[<>|*?/\\]`, name); m {
		return this.Error(`名称中不能包含以下特殊符号：< > | * ? / \`)
	}

	userUuid := request.FormValue("userUuid")
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}
	user = this.userDao.CheckByUuid(userUuid)

	if puuid != "" && puuid != "root" {
		//找出上一级的文件夹。
		this.matterDao.FindByUuidAndUserUuid(puuid, user.Uuid)
	}

	//判断同级文件夹中是否有同名的文件。
	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, puuid, true, name)

	if count > 0 {
		return this.Error("【" + name + "】已经存在了，请使用其他名称。")
	}

	matter := &Matter{
		Puuid:    puuid,
		UserUuid: user.Uuid,
		Dir:      true,
		Name:     name,
	}

	matter = this.matterDao.Create(matter)

	return this.Success(matter)
}

//按照分页的方式获取某个文件夹下文件和子文件夹的列表，通常情况下只有一页。
func (this *MatterController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {

	//如果是根目录，那么就传入root.
	puuid := request.FormValue("puuid")
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	userUuid := request.FormValue("userUuid")
	name := request.FormValue("name")
	dir := request.FormValue("dir")
	orderDir := request.FormValue("orderDir")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderSize := request.FormValue("orderSize")
	orderName := request.FormValue("orderName")
	extensionsStr := request.FormValue("extensions")

	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
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

	//筛选后缀名
	var extensions []string
	if extensionsStr != "" {
		extensions = strings.Split(extensionsStr, ",")
	}

	//文件列表默认文件夹始终在文件的前面。
	if orderDir == "" {
		orderDir = "DESC"
	}

	sortArray := []OrderPair{
		{
			key:   "dir",
			value: orderDir,
		},
		{
			key:   "create_time",
			value: orderCreateTime,
		},
		{
			key:   "size",
			value: orderSize,
		},
		{
			key:   "name",
			value: orderName,
		},
	}

	pager := this.matterDao.Page(page, pageSize, puuid, userUuid, name, dir, extensions, sortArray)

	return this.Success(pager)
}

//上传文件
func (this *MatterController) Upload(writer http.ResponseWriter, request *http.Request) *WebResult {

	userUuid := request.FormValue("userUuid")
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}
	user = this.userDao.CheckByUuid(userUuid)

	alienStr := request.FormValue("alien")
	alien := false
	puuid := ""

	if alienStr == "true" {
		alien = true

		//如果是应用文件的话，统一放在同一个地方。
		puuid = this.matterService.GetDirUuid(userUuid, fmt.Sprintf("/应用数据/%s", time.Now().Local().Format("20060102150405")))

	} else {
		puuid = request.FormValue("puuid")
		if puuid == "" {
			return this.Error("puuid必填")
		} else {
			if puuid != "root" {
				//找出上一级的文件夹。
				this.matterDao.FindByUuidAndUserUuid(puuid, userUuid)

			}
		}
	}

	privacy := false
	privacyStr := request.FormValue("privacy")
	if privacyStr == "true" {
		privacy = true
	}

	request.ParseMultipartForm(32 << 20)
	file, handler, err := request.FormFile("file")
	this.PanicError(err)
	defer file.Close()

	matter := this.matterService.Upload(file, user, puuid, handler.Filename, privacy, alien)

	return this.Success(matter)
}

//删除一个文件
func (this *MatterController) Delete(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		return this.Error("文件的uuid必填")
	}

	matter := this.matterDao.FindByUuid(uuid)

	//判断文件的所属人是否正确
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
		return this.Error(RESULT_CODE_UNAUTHORIZED)
	}

	this.matterDao.Delete(matter)

	return this.Success("删除成功！")
}

//删除一系列文件。
func (this *MatterController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		return this.Error("文件的uuids必填")
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		matter := this.matterDao.FindByUuid(uuid)

		//判断文件的所属人是否正确
		user := this.checkUser(writer, request)
		if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
			return this.Error(RESULT_CODE_UNAUTHORIZED)
		}

		this.matterDao.Delete(matter)

	}

	return this.Success("删除成功！")
}

//重命名一个文件或一个文件夹
func (this *MatterController) Rename(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")
	name := request.FormValue("name")

	//验证参数。
	if name == "" {
		return this.Error("name参数必填")
	}
	if m, _ := regexp.MatchString(`[<>|*?/\\]`, name); m {
		return this.Error(`名称中不能包含以下特殊符号：< > | * ? / \`)
	}

	//找出该文件或者文件夹
	matter := this.matterDao.CheckByUuid(uuid)

	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
		return this.Error(RESULT_CODE_UNAUTHORIZED)
	}

	if name == matter.Name {
		return this.Error("新名称和旧名称一样，操作失败！")
	}

	//判断同级文件夹中是否有同名的文件
	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, matter.Puuid, matter.Dir, name)

	if count > 0 {
		return this.Error("【" + name + "】已经存在了，请使用其他名称。")
	}

	matter.Name = name
	matter = this.matterDao.Save(matter)

	return this.Success(matter)
}

//改变一个文件的公私有属性
func (this *MatterController) ChangePrivacy(writer http.ResponseWriter, request *http.Request) *WebResult {
	uuid := request.FormValue("uuid")
	privacyStr := request.FormValue("privacy")
	privacy := false
	if privacyStr == "true" {
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
		return this.Error(RESULT_CODE_UNAUTHORIZED)
	}

	matter.Privacy = privacy
	this.matterDao.Save(matter)

	return this.Success("设置成功")
}

//将一个文件夹或者文件移入到另一个文件夹下。
func (this *MatterController) Move(writer http.ResponseWriter, request *http.Request) *WebResult {

	srcUuidsStr := request.FormValue("srcUuids")
	destUuid := request.FormValue("destUuid")

	var srcUuids []string
	//验证参数。
	if srcUuidsStr == "" {
		return this.Error("srcUuids参数必填")
	} else {
		srcUuids = strings.Split(srcUuidsStr, ",")
	}

	userUuid := request.FormValue("userUuid")
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}
	if userUuid == "" {
		userUuid = user.Uuid
	}

	user = this.userDao.CheckByUuid(userUuid)

	//验证dest是否有问题
	var destMatter *Matter
	if destUuid == "" {
		return this.Error("destUuid参数必填")
	} else {
		if destUuid != "root" {
			destMatter = this.matterDao.FindByUuid(destUuid)

			if user.Role != USER_ROLE_ADMINISTRATOR && destMatter.UserUuid != user.Uuid {
				return this.Error(RESULT_CODE_UNAUTHORIZED)
			}
		}
	}

	var srcMatters []*Matter
	//验证src是否有问题。
	for _, uuid := range srcUuids {
		//找出该文件或者文件夹
		srcMatter := this.matterDao.FindByUuid(uuid)

		if user.Role != USER_ROLE_ADMINISTRATOR && srcMatter.UserUuid != user.Uuid {
			return this.Error(RESULT_CODE_UNAUTHORIZED)
		}

		if srcMatter.Puuid == destUuid {
			return this.Error("没有进行移动，操作无效！")
		}

		//判断同级文件夹中是否有同名的文件
		count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, destUuid, srcMatter.Dir, srcMatter.Name)

		if count > 0 {
			return this.Error("【" + srcMatter.Name + "】在目标文件夹已经存在了，操作失败。")
		}


		//判断和目标文件夹是否是同一个主人。
		if destUuid != "root" {
			if srcMatter.UserUuid != destMatter.UserUuid {
				panic("文件和目标文件夹的拥有者不是同一人")
			}
		}

		srcMatters = append(srcMatters, srcMatter)
	}

	for _, srcMatter := range srcMatters {
		srcMatter.Puuid = destUuid
		srcMatter = this.matterDao.Save(srcMatter)
	}

	return this.Success(nil)
}
