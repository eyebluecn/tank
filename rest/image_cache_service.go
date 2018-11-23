package rest

import (
	"net/http"
	"image"
	"os"
	"strconv"
	"github.com/disintegration/imaging"
	"net/url"
)

//@Service
type ImageCacheService struct {
	Bean
	imageCacheDao *ImageCacheDao
	userDao       *UserDao
}

//初始化方法
func (this *ImageCacheService) Init(context *Context) {

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := context.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = context.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *ImageCacheService) Detail(uuid string) *ImageCache {

	imageCache := this.imageCacheDao.CheckByUuid(uuid)

	return imageCache
}

//图片预处理功能。
func (this *ImageCacheService) ResizeImage(request *http.Request, filePath string) *image.NRGBA {

	diskFile, err := os.Open(filePath)
	this.PanicError(err)
	defer diskFile.Close()

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
			return imaging.Resize(src, imageResizeW, 0, imaging.Lanczos)

		} else if imageResizeH > 0 {
			//将图缩略成高度为100，宽度按比例处理。
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, 0, imageResizeH, imaging.Lanczos)

		} else {
			panic("单边缩略必须指定imageResizeW或imageResizeH")
		}
	} else if imageResizeM == "fill" {
		//固定宽高，自动裁剪
		if imageResizeW > 0 && imageResizeH > 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Fill(src, imageResizeW, imageResizeH, imaging.Center, imaging.Lanczos)

		} else {
			panic("固定宽高，自动裁剪 必须同时指定imageResizeW和imageResizeH")
		}
	} else if imageResizeM == "fixed" {
		//强制宽高缩略
		if imageResizeW > 0 && imageResizeH > 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, imageResizeW, imageResizeH, imaging.Lanczos)

		} else {
			panic("强制宽高缩略必须同时指定imageResizeW和imageResizeH")
		}
	} else {
		panic("不支持" + imageResizeM + "处理模式")
	}
}

//缓存一张处理完毕了的图片
func (this *ImageCacheService) cacheImage(writer http.ResponseWriter, request *http.Request, matter *Matter) *ImageCache {

	// 防止中文乱码
	fileName := url.QueryEscape(matter.Name)

	//当前的文件是否是图片，只有图片才能处理。
	extension := GetExtension(fileName)
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

	//resize图片
	dstImage := this.ResizeImage(request, CONFIG.MatterPath+matter.Path)

	user := this.userDao.FindByUuid(matter.UserUuid)
	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	absolutePath, relativePath := GetUserFilePath(user.Username, true)
	absolutePath = absolutePath + "/" + fileName
	relativePath = relativePath + "/" + fileName

	fileWriter, err := os.Create(absolutePath)
	this.PanicError(err)
	defer fileWriter.Close()

	//处理后的图片存放在本地
	err = imaging.Encode(fileWriter, dstImage, format)
	this.PanicError(err)

	//获取新文件的大小
	fileInfo, err := fileWriter.Stat()
	this.PanicError(err)

	//相关信息写到缓存中去
	imageCache := &ImageCache{
		UserUuid:   matter.UserUuid,
		MatterUuid: matter.Uuid,
		Uri:        request.RequestURI,
		Size:       fileInfo.Size(),
		Path:       relativePath,
	}
	this.imageCacheDao.Create(imageCache)

	return imageCache
}
