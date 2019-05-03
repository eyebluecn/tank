package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"time"
)

//@Service
type AlienService struct {
	BaseBean
	matterDao         *MatterDao
	matterService     *MatterService
	userDao           *UserDao
	uploadTokenDao    *UploadTokenDao
	downloadTokenDao  *DownloadTokenDao
	shareService      *ShareService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
}

//初始化方法
func (this *AlienService) Init() {
	this.BaseBean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.uploadTokenDao)
	if c, ok := b.(*UploadTokenDao); ok {
		this.uploadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.downloadTokenDao)
	if c, ok := b.(*DownloadTokenDao); ok {
		this.downloadTokenDao = c
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if c, ok := b.(*ShareService); ok {
		this.shareService = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if c, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = c
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
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
				panic(result.UNAUTHORIZED)
			}

			//下载之后立即过期掉。如果是分块下载的，必须以最终获取到完整的数据为准。
			downloadToken.ExpireTime = time.Now()
			this.downloadTokenDao.Save(downloadToken)

		} else {

			//判断文件的所属人是否正确
			operator := this.findUser(request)

			//可以使用分享码的形式授权。
			shareUuid := request.FormValue("shareUuid")
			shareCode := request.FormValue("shareCode")
			shareRootUuid := request.FormValue("shareRootUuid")

			this.shareService.ValidateMatter(shareUuid, shareCode, operator, shareRootUuid, matter)

		}
	}

	//文件夹下载
	if matter.Dir {

		this.logger.Info("准备下载文件夹 %s", matter.Name)

		//目标地点
		this.matterService.DownloadZip(writer, request, []*Matter{matter})

	} else {

		//对图片处理。
		needProcess, imageResizeM, imageResizeW, imageResizeH := this.imageCacheService.ResizeParams(request)
		if needProcess {

			//如果是图片，那么能用缓存就用缓存
			imageCache := this.imageCacheDao.FindByMatterUuidAndMode(matter.Uuid, fmt.Sprintf("%s_%d_%d", imageResizeM, imageResizeW, imageResizeH))
			if imageCache == nil {
				imageCache = this.imageCacheService.cacheImage(writer, request, matter)
			}

			//直接使用缓存中的信息
			this.matterService.DownloadFile(writer, request, GetUserCacheRootDir(imageCache.Username)+imageCache.Path, imageCache.Name, withContentDisposition)

		} else {
			this.matterService.DownloadFile(writer, request, matter.AbsolutePath(), matter.Name, withContentDisposition)
		}

	}

	//文件下载次数加一，为了加快访问速度，异步进行
	go core.RunWithRecovery(func() {
		this.matterDao.TimesIncrement(uuid)
	})

}
