package rest

import (
	"fmt"
	"net/http"
	"time"
)

//@Service
type AlienService struct {
	Bean
	matterDao         *MatterDao
	matterService     *MatterService
	userDao           *UserDao
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
}

//初始化方法
func (this *AlienService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = CONTEXT.GetBean(this.imageCacheService)
	if c, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = c
	}
}

//预览或者下载的统一处理.
func (this *AlienService) PreviewOrDownload(
	writer http.ResponseWriter,
	request *http.Request,
	uuid string,
	filename string,
	withContentDisposition bool) {

	matter := this.matterDao.CheckByUuid(uuid)

	//判断是否是文件夹
	if matter.Dir {
		panic("不支持下载文件夹")
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
				panic(CODE_WRAPPER_UNAUTHORIZED)
			}

			//下载之后立即过期掉。如果是分块下载的，必须以最终获取到完整的数据为准。
			downloadToken.ExpireTime = time.Now()
			this.downloadTokenDao.Save(downloadToken)

		} else {

			//判断文件的所属人是否正确
			operator := this.findUser(writer, request)
			if operator == nil || (operator.Role != USER_ROLE_ADMINISTRATOR && matter.UserUuid != operator.Uuid) {
				panic(CODE_WRAPPER_UNAUTHORIZED)
			}

		}
	}

	//对图片处理。
	needProcess, imageResizeM, imageResizeW, imageResizeH := this.imageCacheService.ResizeParams(request)
	if needProcess {

		//如果是图片，那么能用缓存就用缓存
		imageCache := this.imageCacheDao.FindByMatterUuidAndMode(matter.Uuid, fmt.Sprintf("%s_%d_%d", imageResizeM, imageResizeW, imageResizeH))
		if imageCache == nil {
			imageCache = this.imageCacheService.cacheImage(writer, request, matter)
		}

		//直接使用缓存中的信息
		this.matterService.DownloadFile(writer, request, CONFIG.MatterPath+imageCache.Path, matter.Name, withContentDisposition)

	} else {
		this.matterService.DownloadFile(writer, request, CONFIG.MatterPath+matter.Path, matter.Name, withContentDisposition)
	}

	//文件下载次数加一，为了加快访问速度，异步进行
	go SafeMethod(func() {
		this.matterDao.TimesIncrement(uuid)
	})

}
