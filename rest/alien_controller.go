package rest

import (
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AlienController struct {
	BaseController
	uploadTokenDao   *UploadTokenDao
	downloadTokenDao *DownloadTokenDao
	matterDao        *MatterDao
	matterService    *MatterService
}

//初始化方法
func (this *AlienController) Init(context *Context) {
	this.BaseController.Init(context)

	//手动装填本实例的Bean.
	b := context.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = context.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = context.GetBean(this.matterDao)
	if c, ok := b.(*MatterDao); ok {
		this.matterDao = c
	}

	b = context.GetBean(this.matterService)
	if c, ok := b.(*MatterService); ok {
		this.matterService = c
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

	return routeMap
}

//处理一些特殊的接口，比如参数包含在路径中,一般情况下，controller不将参数放在url路径中
func (this *AlienController) HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool) {

	path := request.URL.Path

	//匹配 /api/alien/download/{uuid}/{filename}
	reg := regexp.MustCompile(`^/api/alien/download/([^/]+)/([^/]+)$`)
	strs := reg.FindStringSubmatch(path)
	if len(strs) != 3 {
		return nil, false
	} else {
		var f = func(writer http.ResponseWriter, request *http.Request) {
			this.Download(writer, request, strs[1], strs[2])
		}
		return f, true
	}
}

//直接使用邮箱和密码获取用户
func (this *AlienController) CheckRequestUser(email, password string) *User {

	if email == "" {
		panic("邮箱必填啦")
	}

	if password == "" {
		panic("密码必填")
	}

	//验证用户身份合法性。
	user := this.userDao.FindByEmail(email)
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

	user := this.CheckRequestUser(request.FormValue("email"), request.FormValue("password"))
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

	user := this.CheckRequestUser(request.FormValue("email"), request.FormValue("password"))

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

//系统中的用户x要获取一个DownloadToken，用于提供给x信任的用户下载文件。
func (this *AlienController) FetchDownloadToken(writer http.ResponseWriter, request *http.Request) *WebResult {

	matterUuid := request.FormValue("matterUuid")
	if matterUuid == "" {
		panic("matterUuid必填")
	}

	user := this.CheckRequestUser(request.FormValue("email"), request.FormValue("password"))

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

//下载一个文件。既可以使用登录的方式下载，也可以使用授权的方式下载。
func (this *AlienController) Download(writer http.ResponseWriter, request *http.Request, uuid string, filename string) {

	matter := this.matterDao.CheckByUuid(uuid)

	//判断是否是文件夹
	if matter.Dir {
		panic("暂不支持下载文件夹")
	}

	if matter.Name != filename {
		panic("文件信息错误")
	}

	//验证用户的权限问题。
	//文件如果是私有的才需要权限
	if matter.Privacy {

		//1.如果带有downloadTokenUuid那么就按照token的信息去获取。
		downloadTokenUuid := request.FormValue("downloadTokenUuid")
		if downloadTokenUuid != "" {
			downloadToken := this.downloadTokenDao.CheckByUuid(downloadTokenUuid)
			if downloadToken.ExpireTime.Before(time.Now()) {
				panic("downloadToken已失效")
			}

			if downloadToken.MatterUuid != uuid {
				panic("token和文件信息不一致")
			}

			tokenUser := this.userDao.CheckByUuid(downloadToken.UserUuid)
			if matter.UserUuid != tokenUser.Uuid {
				panic(RESULT_CODE_UNAUTHORIZED)
			}

			//下载之后立即过期掉。
			downloadToken.ExpireTime = time.Now()
			this.downloadTokenDao.Save(downloadToken)

		} else {

			//判断文件的所属人是否正确
			user := this.checkUser(writer, request)
			if user.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != user.Uuid {
				panic(RESULT_CODE_UNAUTHORIZED)
			}

		}
	}

	diskFile, err := os.Open(GetFilePath() + matter.Path)
	this.PanicError(err)
	defer diskFile.Close()
	// 防止中文乱码
	fileName := url.QueryEscape(matter.Name)
	writer.Header().Set("Content-Type", GetMimeType(fileName))

	//如果是图片或者文本就直接打开。
	mimeType := GetMimeType(matter.Name)
	if strings.Index(mimeType, "image") != 0 && strings.Index(mimeType, "text") != 0 {
		writer.Header().Set("content-disposition", "attachment; filename=\""+fileName+"\"")
	}

	//对图片做缩放处理。
	imageProcess := request.FormValue("imageProcess")
	if imageProcess == "resize" {

		//当前的文件是否是图片，只有图片才能处理。
		extension := GetExtension(matter.Name)
		formats := map[string]imaging.Format{
			".jpg":  imaging.JPEG,
			".jpeg": imaging.JPEG,
			".png":  imaging.PNG,
			".tif":  imaging.TIFF,
			".tiff": imaging.TIFF,
			".bmp":  imaging.BMP,
			".gif":  imaging.GIF,
		}

		format, ok := formats[extension]
		if !ok {
			panic("该图片格式不支持处理")
		}

		imageResizeM := request.FormValue("imageResizeM")
		if imageResizeM == "" {
			imageResizeM = "fit"
		} else if imageResizeM != "fit" && imageResizeM != "fill" && imageResizeM != "fixed" {
			panic("imageResizeM参数错误")
		}
		imageResizeWStr := request.FormValue("imageResizeW")
		var imageResizeW int
		if imageResizeWStr != "" {
			imageResizeW, err = strconv.Atoi(imageResizeWStr)
			this.PanicError(err)
			if imageResizeW < 1 || imageResizeW > 4096 {
				panic("缩放尺寸不能超过4096")
			}
		}
		imageResizeHStr := request.FormValue("imageResizeH")
		var imageResizeH int
		if imageResizeHStr != "" {
			imageResizeH, err = strconv.Atoi(imageResizeHStr)
			this.PanicError(err)
			if imageResizeH < 1 || imageResizeH > 4096 {
				panic("缩放尺寸不能超过4096")
			}
		}

		//单边缩略
		if imageResizeM == "fit" {
			//将图缩略成宽度为100，高度按比例处理。
			if imageResizeW > 0 {
				src, err := imaging.Decode(diskFile)
				this.PanicError(err)
				dst := imaging.Resize(src, imageResizeW, 0, imaging.Lanczos)

				err = imaging.Encode(writer, dst, format)
				this.PanicError(err)
			} else if imageResizeH > 0 {
				//将图缩略成高度为100，宽度按比例处理。
				src, err := imaging.Decode(diskFile)
				this.PanicError(err)
				dst := imaging.Resize(src, 0, imageResizeH, imaging.Lanczos)

				err = imaging.Encode(writer, dst, format)
				this.PanicError(err)
			} else {
				panic("单边缩略必须指定imageResizeW或imageResizeH")
			}
		} else if imageResizeM == "fill" {
			//固定宽高，自动裁剪
			if imageResizeW > 0 && imageResizeH > 0 {
				src, err := imaging.Decode(diskFile)
				this.PanicError(err)
				dst := imaging.Fill(src, imageResizeW, imageResizeH, imaging.Center, imaging.Lanczos)
				err = imaging.Encode(writer, dst, format)
				this.PanicError(err)
			} else {
				panic("固定宽高，自动裁剪 必须同时指定imageResizeW和imageResizeH")
			}
		} else if imageResizeM == "fixed" {
			//强制宽高缩略
			if imageResizeW > 0 && imageResizeH > 0 {
				src, err := imaging.Decode(diskFile)
				this.PanicError(err)
				dst := imaging.Resize(src, imageResizeW, imageResizeH, imaging.Lanczos)

				err = imaging.Encode(writer, dst, format)
				this.PanicError(err)
			} else {
				panic("强制宽高缩略必须同时指定imageResizeW和imageResizeH")
			}
		}
	} else {

		_, err = io.Copy(writer, diskFile)
		this.PanicError(err)

	}

}
