package rest

import (
	"io"
	"mime/multipart"
	"os"
	"regexp"
	"strings"
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

	puuid := "root"
	for k, name := range folders {
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

//开始上传文件
//上传文件. alien表明文件是否是应用使用的文件。
func (this *MatterService) Upload(file multipart.File, user *User, puuid string, filename string, privacy bool, alien bool) *Matter {

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
