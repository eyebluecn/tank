package rest

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// @Service
type ImageCacheService struct {
	BaseBean
	imageCacheDao *ImageCacheDao
	userDao       *UserDao
	matterDao     *MatterDao
}

func (this *ImageCacheService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

}

func (this *ImageCacheService) Detail(uuid string) *ImageCache {

	imageCache := this.imageCacheDao.CheckByUuid(uuid)

	return imageCache
}

// prepare the resize parameters.
func (this *ImageCacheService) ResizeParams(request *http.Request) (needProcess bool, resizeMode string, resizeWidth int, resizeHeight int) {
	var err error

	if request.FormValue("ir") != "" {
		//mode_w_h  if w or h equal means not required.
		imageResizeStr := request.FormValue("ir")
		arr := strings.Split(imageResizeStr, "_")
		if len(arr) != 3 {
			panic(result.BadRequest("param error. the format is mode_w_h"))
		}

		imageResizeM := arr[0]
		if imageResizeM == "" {
			imageResizeM = "fit"
		} else if imageResizeM != "fit" && imageResizeM != "fill" && imageResizeM != "fixed" {
			panic(result.BadRequest("mode can only be fit/fill/fixed"))
		}
		imageResizeWStr := arr[1]
		var imageResizeW int
		if imageResizeWStr != "" {
			imageResizeW, err = strconv.Atoi(imageResizeWStr)
			this.PanicError(err)
			if imageResizeW < 0 || imageResizeW > 4096 {
				panic(result.BadRequest("zoom size cannot exceed 4096"))
			}
		}
		imageResizeHStr := arr[2]
		var imageResizeH int
		if imageResizeHStr != "" {
			imageResizeH, err = strconv.Atoi(imageResizeHStr)
			this.PanicError(err)
			if imageResizeH < 0 || imageResizeH > 4096 {
				panic(result.BadRequest("zoom size cannot exceed 4096"))
			}
		}
		return true, imageResizeM, imageResizeW, imageResizeH
	} else {
		return false, "", 0, 0
	}

}

// resize image.
func (this *ImageCacheService) ResizeImage(request *http.Request, filePath string) *image.NRGBA {

	diskFile, err := os.Open(filePath)
	this.PanicError(err)
	defer func() {
		e := diskFile.Close()
		this.PanicError(e)
	}()

	_, imageResizeM, imageResizeW, imageResizeH := this.ResizeParams(request)

	if imageResizeM == "fit" {
		//fit mode.
		if imageResizeW != 0 {
			//eg. width = 100 height auto in proportion

			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, imageResizeW, 0, imaging.Lanczos)

		} else if imageResizeH != 0 {
			//eg. height = 100 width auto in proportion

			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, 0, imageResizeH, imaging.Lanczos)

		} else {
			panic(result.BadRequest("mode fit required width or height"))
		}
	} else if imageResizeM == "fill" {
		//fill mode. specify the width and height
		if imageResizeW > 0 && imageResizeH > 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Fill(src, imageResizeW, imageResizeH, imaging.Center, imaging.Lanczos)

		} else {
			panic(result.BadRequest("mode fill required width and height"))
		}
	} else if imageResizeM == "fixed" {
		//fixed mode
		if imageResizeW > 0 && imageResizeH > 0 {
			src, err := imaging.Decode(diskFile)
			this.PanicError(err)
			return imaging.Resize(src, imageResizeW, imageResizeH, imaging.Lanczos)

		} else {
			panic(result.BadRequest("mode fixed required width and height"))
		}
	} else {
		panic(result.BadRequest("not support mode %s", imageResizeM))
	}
}

// cache an image
func (this *ImageCacheService) cacheImage(writer http.ResponseWriter, request *http.Request, matter *Matter) *ImageCache {

	//only these image can do.
	extension := util.GetExtension(matter.Name)
	formats := map[string]imaging.Format{
		".jpg":  imaging.JPEG,
		".jpeg": imaging.JPEG,
		".png":  imaging.PNG,
		".tif":  imaging.TIFF,
		".tiff": imaging.TIFF,
		".bmp":  imaging.BMP,
		".gif":  imaging.GIF,
	}

	_, imageResizeM, imageResizeW, imageResizeH := this.ResizeParams(request)
	mode := fmt.Sprintf("%s_%d_%d", imageResizeM, imageResizeW, imageResizeH)

	format, ok := formats[extension]
	if !ok {
		panic(result.BadRequest("not support this kind of image's (%s) resize", extension))
	}

	user := this.userDao.FindByUuid(matter.UserUuid)

	dstImage := this.ResizeImage(request, matter.AbsolutePath())

	cacheImageName := util.GetSimpleFileName(matter.Name) + "_" + mode + extension
	cacheImageRelativePath := util.GetSimpleFileName(matter.Path) + "_" + mode + extension
	cacheImageAbsolutePath := GetSpaceCacheRootDir(user.Username) + util.GetSimpleFileName(matter.Path) + "_" + mode + extension

	//create directory
	dir := filepath.Dir(cacheImageAbsolutePath)
	util.MakeDirAll(dir)

	fileWriter, err := os.Create(cacheImageAbsolutePath)
	this.PanicError(err)
	defer func() {
		e := fileWriter.Close()
		this.PanicError(e)
	}()

	//store on disk after handle
	err = imaging.Encode(fileWriter, dstImage, format)
	this.PanicError(err)

	fileInfo, err := fileWriter.Stat()
	this.PanicError(err)

	imageCache := &ImageCache{
		Name:       cacheImageName,
		UserUuid:   matter.UserUuid,
		Username:   user.Username,
		MatterUuid: matter.Uuid,
		MatterName: matter.Name,
		Mode:       mode,
		Size:       fileInfo.Size(),
		Path:       cacheImageRelativePath,
	}
	this.imageCacheDao.Create(imageCache)

	return imageCache
}
