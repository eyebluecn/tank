package rest

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//@Service
type ImageCacheService struct {
	Bean
	imageCacheDao *ImageCacheDao
	userDao       *UserDao
}

//初始化方法
func (this *ImageCacheService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *ImageCacheService) Detail(uuid string) *ImageCache {

	imageCache := this.imageCacheDao.CheckByUuid(uuid)

	return imageCache
}

//获取预处理时必要的参数
func (this *ImageCacheService) ResizeParams(request *http.Request) (needProcess bool, resizeMode string, resizeWidth int, resizeHeight int) {
	var err error

	//1.0 模式准备逐步废弃掉
	if request.FormValue("imageProcess") == "resize" {
		//老模式使用 imageResizeM,imageResizeW,imageResizeH
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

		return true, imageResizeM, imageResizeW, imageResizeH
	} else if request.FormValue("ir") != "" {
		//新模式使用 mode_w_h  如果w或者h为0表示这项值不设置
		imageResizeStr := request.FormValue("ir")
		arr := strings.Split(imageResizeStr, "_")
		if len(arr) != 3 {
			panic("参数不符合规范，格式要求为mode_w_h")
		}

		imageResizeM := arr[0]
		if imageResizeM == "" {
			imageResizeM = "fit"
		} else if imageResizeM != "fit" && imageResizeM != "fill" && imageResizeM != "fixed" {
			panic("imageResizeM参数错误")
		}
		imageResizeWStr := arr[1]
		var imageResizeW int
		if imageResizeWStr != "" {
			imageResizeW, err = strconv.Atoi(imageResizeWStr)
			this.PanicError(err)
			if imageResizeW < 0 || imageResizeW > 4096 {
				panic("缩放尺寸不能超过4096")
			}
		}
		imageResizeHStr := arr[2]
		var imageResizeH int
		if imageResizeHStr != "" {
			imageResizeH, err = strconv.Atoi(imageResizeHStr)
			this.PanicError(err)
			if imageResizeH < 0 || imageResizeH > 4096 {
				panic("缩放尺寸不能超过4096")
			}
		}
		return true, imageResizeM, imageResizeW, imageResizeH
	} else {
		return false, "", 0, 0
	}

}

//图片预处理功能。
func (this *ImageCacheService) ResizeImage(request *http.Request, filePath string) *image.NRGBA {

	diskFile, err := os.Open(filePath)
	this.PanicError(err)
	defer diskFile.Close()

	_, imageResizeM, imageResizeW, imageResizeH := this.ResizeParams(request)

	//单边缩略
	if imageResizeM == "fit" {
		//将图缩略成宽度为100，高度按比例处理。
		if imageResizeW != 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, imageResizeW, 0, imaging.Lanczos)

		} else if imageResizeH != 0 {
			//将图缩略成高度为100，宽度按比例处理。
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, 0, imageResizeH, imaging.Lanczos)

		} else {
			panic("单边缩略必须指定宽或者高")
		}
	} else if imageResizeM == "fill" {
		//固定宽高，自动裁剪
		if imageResizeW > 0 && imageResizeH > 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Fill(src, imageResizeW, imageResizeH, imaging.Center, imaging.Lanczos)

		} else {
			panic("固定宽高，自动裁剪 必须同时指定宽和高")
		}
	} else if imageResizeM == "fixed" {
		//强制宽高缩略
		if imageResizeW > 0 && imageResizeH > 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, imageResizeW, imageResizeH, imaging.Lanczos)

		} else {
			panic("强制宽高缩略必须同时指定宽和高")
		}
	} else {
		panic("不支持" + imageResizeM + "处理模式")
	}
}

//缓存一张处理完毕了的图片
func (this *ImageCacheService) cacheImage(writer http.ResponseWriter, request *http.Request, matter *Matter) *ImageCache {

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

	//resize图片
	dstImage := this.ResizeImage(request, CONFIG.MatterPath+matter.Path)

	user := this.userDao.FindByUuid(matter.UserUuid)
	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	absolutePath, relativePath := GetUserFilePath(user.Username, true)
	absolutePath = absolutePath + "/" + matter.Name
	relativePath = relativePath + "/" + matter.Name

	fileWriter, err := os.Create(absolutePath)
	this.PanicError(err)
	defer fileWriter.Close()

	//处理后的图片存放在本地
	err = imaging.Encode(fileWriter, dstImage, format)
	this.PanicError(err)

	//获取新文件的大小
	fileInfo, err := fileWriter.Stat()
	this.PanicError(err)

	_, imageResizeM, imageResizeW, imageResizeH := this.ResizeParams(request)

	//相关信息写到缓存中去
	imageCache := &ImageCache{
		UserUuid:   matter.UserUuid,
		MatterUuid: matter.Uuid,
		Mode:       fmt.Sprintf("%s_%d_%d", imageResizeM, imageResizeW, imageResizeH),
		Size:       fileInfo.Size(),
		Path:       relativePath,
	}
	this.imageCacheDao.Create(imageCache)

	return imageCache
}
