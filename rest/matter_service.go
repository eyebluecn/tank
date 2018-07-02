package rest

import (
	"io"
	"mime/multipart"
	"os"
	"regexp"
	"strings"
	"net/http"
	"github.com/disintegration/imaging"
	"strconv"
)

//@Service
type MatterService struct {
	Bean
	matterDao *MatterDao
}

//初始化方法
func (this *MatterService) Init(context *Context) {

	//手动装填本实例的Bean. 这里必须要用中间变量方可。

	b := context.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

}

//根据一个文件夹路径，找到最后一个文件夹的uuid，如果中途出错，返回err.
func (this *MatterService) GetDirUuid(userUuid string, dir string) string {

	if dir == "" {
		panic(`文件夹不能为空`)
	} else if dir[0:1] != "/" {
		panic(`文件夹必须以/开头`)
	} else if strings.Index(dir, "//") != -1 {
		panic(`文件夹不能出现连续的//`)
	} else if m, _ := regexp.MatchString(`[<>|*?\\]`, dir); m {
		panic(`文件夹中不能包含以下特殊符号：< > | * ? \`)
	}

	if dir == "/" {
		return "root"
	}

	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}

	//递归找寻文件的上级目录uuid.
	folders := strings.Split(dir, "/")

	if len(folders) > 32 {
		panic("文件夹最多32层。")
	}

	puuid := "root"
	for k, name := range folders {

		if len(name) > 200 {
			panic("每级文件夹的最大长度为200")
		}

		if k == 0 {
			continue
		}

		matter := this.matterDao.FindByUserUuidAndPuuidAndNameAndDirTrue(userUuid, puuid, name)
		if matter == nil {
			//创建一个文件夹。这里一般都是通过alien接口来创建的文件夹。
			matter = &Matter{
				Puuid:    puuid,
				UserUuid: userUuid,
				Dir:      true,
				Alien:    true,
				Name:     name,
			}
			matter = this.matterDao.Create(matter)
		}

		puuid = matter.Uuid
	}

	return puuid
}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *MatterService) Detail(uuid string) *Matter {

	matter := this.matterDao.CheckByUuid(uuid)

	//组装file的内容，展示其父组件。
	puuid := matter.Puuid
	tmpMatter := matter
	for puuid != "root" {
		pFile := this.matterDao.CheckByUuid(puuid)
		tmpMatter.Parent = pFile
		tmpMatter = pFile
		puuid = pFile.Puuid
	}

	return matter
}

//开始上传文件
//上传文件. alien表明文件是否是应用使用的文件。
func (this *MatterService) Upload(file multipart.File, user *User, puuid string, filename string, privacy bool, alien bool) *Matter {

	//文件名不能太长。
	if len(filename) > 200 {
		panic("文件名不能超过200")
	}

	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	absolutePath, relativePath := GetUserFilePath(user.Username)
	absolutePath = absolutePath + "/" + filename
	relativePath = relativePath + "/" + filename

	distFile, err := os.OpenFile(absolutePath, os.O_WRONLY|os.O_CREATE, 0777)
	this.PanicError(err)

	defer distFile.Close()

	written, err := io.Copy(distFile, file)
	this.PanicError(err)

	//判断用户自身上传大小的限制。
	if user.SizeLimit >= 0 {
		if written > user.SizeLimit {
			panic("您最大只能上传" + HumanFileSize(user.SizeLimit) + "的文件")
		}
	}

	//查找文件夹下面是否有同名文件。
	matters := this.matterDao.ListByUserUuidAndPuuidAndDirAndName(user.Uuid, puuid, false, filename)
	//如果有同名的文件，那么我们直接覆盖同名文件。
	for _, dbFile := range matters {
		this.matterDao.Delete(dbFile)
	}

	//将文件信息存入数据库中。
	matter := &Matter{
		Puuid:    puuid,
		UserUuid: user.Uuid,
		Dir:      false,
		Alien:    alien,
		Name:     filename,
		Md5:      "",
		Size:     written,
		Privacy:  privacy,
		Path:     relativePath,
	}

	matter = this.matterDao.Create(matter)

	return matter
}

//处理图片下载功能。
func (this *MatterService) ResizeImage(writer http.ResponseWriter, request *http.Request, matter *Matter, diskFile *os.File) {

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
		imageResizeW, err := strconv.Atoi(imageResizeWStr)
		this.PanicError(err)
		if imageResizeW < 1 || imageResizeW > 4096 {
			panic("缩放尺寸不能超过4096")
		}
	}
	imageResizeHStr := request.FormValue("imageResizeH")
	var imageResizeH int
	if imageResizeHStr != "" {
		imageResizeH, err := strconv.Atoi(imageResizeHStr)
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

}
