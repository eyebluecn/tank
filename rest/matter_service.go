package rest

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"tank/rest/download"
	"tank/rest/result"
)

//@Service
type MatterService struct {
	Bean
	matterDao         *MatterDao
	userDao           *UserDao
	userService       *UserService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
}

//初始化方法
func (this *MatterService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
	}

	b = CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = CONTEXT.GetBean(this.imageCacheService)
	if b, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = b
	}

}

//文件下载。支持分片下载
func (this *MatterService) DownloadFile(
	writer http.ResponseWriter,
	request *http.Request,
	filePath string,
	filename string,
	withContentDisposition bool) {

	download.DownloadFile(writer, request, filePath, filename, withContentDisposition)
}

//删除文件
func (this *MatterService) Delete(matter *Matter) {

	if matter == nil {
		panic(result.BadRequest("matter不能为nil"))
	}



	//操作锁
	this.userService.MatterLock(matter.UserUuid)
	defer this.userService.MatterUnlock(matter.UserUuid)



	this.matterDao.Delete(matter)

}


//开始上传文件
//上传文件. alien表明文件是否是应用使用的文件。
func (this *MatterService) Upload(file io.Reader, user *User, puuid string, filename string, privacy bool, alien bool) *Matter {

	//文件名不能太长。
	if len(filename) > 200 {
		panic("文件名不能超过200")
	}

	//文件夹路径
	var dirAbsolutePath string
	var dirRelativePath string
	if puuid == "" {
		this.PanicBadRequest("puuid必填")
	} else {

		if puuid == MATTER_ROOT {
			dirAbsolutePath = GetUserFileRootDir(user.Username)
			dirRelativePath = ""
		} else {
			//验证puuid是否存在
			dirMatter := this.matterDao.CheckByUuidAndUserUuid(puuid, user.Uuid)

			dirAbsolutePath = GetUserFileRootDir(user.Username) + dirMatter.Path
			dirRelativePath = dirMatter.Path

		}
	}

	//查找文件夹下面是否有同名文件。
	matters := this.matterDao.ListByUserUuidAndPuuidAndDirAndName(user.Uuid, puuid, false, filename)
	//如果有同名的文件，那么我们直接覆盖同名文件。
	for _, dbFile := range matters {
		this.PanicBadRequest("该目录下%s已经存在了", dbFile.Name)
	}

	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	fileAbsolutePath := dirAbsolutePath + "/" + filename
	fileRelativePath := dirRelativePath + "/" + filename

	//创建父文件夹
	MakeDirAll(dirAbsolutePath)

	//如果文件已经存在了，那么直接覆盖。
	exist, err := PathExists(fileAbsolutePath)
	this.PanicError(err)
	if exist {
		this.logger.Error("%s已经存在，将其删除", fileAbsolutePath)
		removeError := os.Remove(fileAbsolutePath)
		this.PanicError(removeError)
	}

	distFile, err := os.OpenFile(fileAbsolutePath, os.O_WRONLY|os.O_CREATE, 0777)
	this.PanicError(err)

	defer func() {
		err := distFile.Close()
		this.PanicError(err)
	}()

	written, err := io.Copy(distFile, file)
	this.PanicError(err)

	this.logger.Info("上传文件%s大小为%v", filename, HumanFileSize(written))

	//判断用户自身上传大小的限制。
	if user.SizeLimit >= 0 {
		if written > user.SizeLimit {
			this.PanicBadRequest("文件大小超出限制 " + HumanFileSize(user.SizeLimit) + ">" + HumanFileSize(written))
		}
	}

	//将文件信息存入数据库中。
	matter := &Matter{
		Puuid:    puuid,
		UserUuid: user.Uuid,
		Username: user.Username,
		Dir:      false,
		Alien:    alien,
		Name:     filename,
		Md5:      "",
		Size:     written,
		Privacy:  privacy,
		Path:     fileRelativePath,
	}

	matter = this.matterDao.Create(matter)

	return matter
}



//根据一个文件夹路径，找到最后一个文件夹的uuid，如果中途出错，返回err.
func (this *MatterService) GetDirUuid(user *User, dir string) (puuid string, dirRelativePath string) {

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
		return MATTER_ROOT, ""
	}

	if dir[len(dir)-1] == '/' {
		dir = dir[:len(dir)-1]
	}

	//递归找寻文件的上级目录uuid.
	folders := strings.Split(dir, "/")

	if len(folders) > 32 {
		panic("文件夹最多32层。")
	}

	puuid = MATTER_ROOT
	parentRelativePath := "/"
	for k, name := range folders {

		if len(name) > 200 {
			panic("每级文件夹的最大长度为200")
		}

		if k == 0 {
			continue
		}

		matter := this.matterDao.FindByUserUuidAndPuuidAndNameAndDirTrue(user.Uuid, puuid, name)
		if matter == nil {
			//创建一个文件夹。这里一般都是通过alien接口来创建的文件夹。
			matter = &Matter{
				Puuid:    puuid,
				UserUuid: user.Uuid,
				Username: user.Username,
				Dir:      true,
				Alien:    true,
				Name:     name,
				Path:     parentRelativePath + "/" + name,
			}
			matter = this.matterDao.Create(matter)
		}

		puuid = matter.Uuid
		parentRelativePath = matter.Path
	}

	return puuid, parentRelativePath
}

//在dirMatter中创建文件夹 返回刚刚创建的这个文件夹
func (this *MatterService) CreateDirectory(dirMatter *Matter, name string, user *User) *Matter {

	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	//父级matter必须存在
	if dirMatter == nil {
		panic(result.BadRequest("dirMatter必须指定"))
	}

	//必须是文件夹
	if !dirMatter.Dir {
		panic(result.BadRequest("dirMatter必须是文件夹"))
	}

	if dirMatter.UserUuid != user.Uuid {

		panic(result.BadRequest("dirMatter的userUuid和user不一致"))
	}

	name = strings.TrimSpace(name)
	//验证参数。
	if name == "" {
		panic(result.BadRequest("name参数必填，并且不能全是空格"))
	}

	if len(name) > MATTER_NAME_MAX_LENGTH {

		panic(result.BadRequest("name长度不能超过%d", MATTER_NAME_MAX_LENGTH))

	}

	if m, _ := regexp.MatchString(`[<>|*?/\\]`, name); m {

		panic(result.BadRequest(`名称中不能包含以下特殊符号：< > | * ? / \`))
	}

	//判断同级文件夹中是否有同名的文件夹
	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, dirMatter.Uuid, true, name)

	if count > 0 {

		panic(result.BadRequest("%s 已经存在了，请使用其他名称。", name))
	}

	parts := strings.Split(dirMatter.Path, "/")
	this.logger.Info("%s的层数：%d", dirMatter.Name, len(parts))

	if len(parts) >= 32 {
		panic(result.BadRequest("文件夹最多%d层", MATTER_NAME_MAX_DEPTH))
	}

	//绝对路径
	absolutePath := GetUserFileRootDir(user.Username) + dirMatter.Path + "/" + name

	//相对路径
	relativePath := dirMatter.Path + "/" + name

	//磁盘中创建文件夹。
	dirPath := MakeDirAll(absolutePath)
	this.logger.Info("Create Directory: %s", dirPath)

	//数据库中创建文件夹。
	matter := &Matter{
		Puuid:    dirMatter.Uuid,
		UserUuid: user.Uuid,
		Username: user.Username,
		Dir:      true,
		Name:     name,
		Path:     relativePath,
	}

	matter = this.matterDao.Create(matter)

	return matter
}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *MatterService) Detail(uuid string) *Matter {

	matter := this.matterDao.CheckByUuid(uuid)

	//组装file的内容，展示其父组件。
	puuid := matter.Puuid
	tmpMatter := matter
	for puuid != MATTER_ROOT {
		pFile := this.matterDao.CheckByUuid(puuid)
		tmpMatter.Parent = pFile
		tmpMatter = pFile
		puuid = pFile.Puuid
	}

	return matter
}

// 从指定的url下载一个文件。参考：https://golangcode.com/download-a-file-from-a-url/
func (this *MatterService) httpDownloadFile(filepath string, url string) (int64, error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return 0, err
	}
	defer func() {
		e := out.Close()
		this.PanicError(e)
	}()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer func() {
		e := resp.Body.Close()
		this.PanicError(e)
	}()

	// Write the body to file
	size, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, err
	}

	return size, nil
}

//去指定的url中爬文件
func (this *MatterService) Crawl(url string, filename string, user *User, puuid string, dirRelativePath string, privacy bool) *Matter {

	//文件名不能太长。
	if len(filename) > 200 {
		panic("文件名不能超过200")
	}

	//获取文件应该存放在的物理路径的绝对路径和相对路径。
	absolutePath := GetUserFileRootDir(user.Username) + dirRelativePath + "/" + filename
	relativePath := dirRelativePath + "/" + filename

	//使用临时文件存放
	fmt.Printf("存放于%s", absolutePath)
	size, err := this.httpDownloadFile(absolutePath, url)
	this.PanicError(err)

	//判断用户自身上传大小的限制。
	if user.SizeLimit >= 0 {
		if size > user.SizeLimit {
			panic("您最大只能上传" + HumanFileSize(user.SizeLimit) + "的文件")
		}
	}

	//查找文件夹下面是否有同名文件。
	matters := this.matterDao.ListByUserUuidAndPuuidAndDirAndName(user.Uuid, puuid, false, filename)
	//如果有同名的文件，那么我们直接覆盖同名文件。
	for _, dbFile := range matters {
		this.Delete(dbFile)
	}

	//将文件信息存入数据库中。
	matter := &Matter{
		Puuid:    puuid,
		UserUuid: user.Uuid,
		Username: user.Username,
		Dir:      false,
		Alien:    false,
		Name:     filename,
		Md5:      "",
		Size:     size,
		Privacy:  privacy,
		Path:     relativePath,
	}

	matter = this.matterDao.Create(matter)

	return matter
}

//调整一个Matter的path值
func (this *MatterService) adjustPath(matter *Matter, parentMatter *Matter) {

	if matter.Dir {
		//如果源是文件夹

		//首先调整好自己
		matter.Path = parentMatter.Path + "/" + matter.Name
		matter = this.matterDao.Save(matter)

		//调整该文件夹下文件的Path.
		matters := this.matterDao.List(matter.Uuid, matter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, matter)
		}

	} else {
		//如果源是普通文件

		//删除该文件的所有缓存
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		//调整path
		matter.Path = parentMatter.Path + "/" + matter.Name
		matter = this.matterDao.Save(matter)
	}

}

//将一个srcMatter放置到另一个destMatter(必须为文件夹)下
func (this *MatterService) Move(srcMatter *Matter, destMatter *Matter) {

	if !destMatter.Dir {
		this.PanicBadRequest("目标必须为文件夹")
	}

	if srcMatter.Dir {
		//如果源是文件夹
		destAbsolutePath := destMatter.AbsolutePath() + "/" + srcMatter.Name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//物理文件一口气移动
		err := os.Rename(srcAbsolutePath, destAbsolutePath)
		this.PanicError(err)

		//修改数据库中信息
		srcMatter.Puuid = destMatter.Uuid
		srcMatter.Path = destMatter.Path + "/" + srcMatter.Name
		srcMatter = this.matterDao.Save(srcMatter)

		//调整该文件夹下文件的Path.
		matters := this.matterDao.List(srcMatter.Uuid, srcMatter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, srcMatter)
		}

	} else {
		//如果源是普通文件

		destAbsolutePath := destMatter.AbsolutePath() + "/" + srcMatter.Name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//物理文件进行移动
		err := os.Rename(srcAbsolutePath, destAbsolutePath)
		this.PanicError(err)

		//删除对应的缓存。
		this.imageCacheDao.DeleteByMatterUuid(srcMatter.Uuid)

		//修改数据库中信息
		srcMatter.Puuid = destMatter.Uuid
		srcMatter.Path = destMatter.Path + "/" + srcMatter.Name
		srcMatter = this.matterDao.Save(srcMatter)

	}

	return
}

//将一个srcMatter复制到另一个destMatter(必须为文件夹)下，名字叫做name
func (this *MatterService) Copy(srcMatter *Matter, destDirMatter *Matter, name string) {

	if !destDirMatter.Dir {
		this.PanicBadRequest("目标必须为文件夹")
	}

	if srcMatter.Dir {

		//如果源是文件夹

		//在目标地址创建新文件夹。
		newMatter := &Matter{
			Puuid:    destDirMatter.Uuid,
			UserUuid: srcMatter.UserUuid,
			Username: srcMatter.Username,
			Dir:      srcMatter.Dir,
			Alien:    srcMatter.Alien,
			Name:     name,
			Md5:      "",
			Size:     srcMatter.Size,
			Privacy:  srcMatter.Privacy,
			Path:     destDirMatter.Path + "/" + name,
		}

		newMatter = this.matterDao.Create(newMatter)

		//复制子文件或文件夹
		matters := this.matterDao.List(srcMatter.Uuid, srcMatter.UserUuid, nil)
		for _, m := range matters {
			this.Copy(m, newMatter, m.Name)
		}

	} else {
		//如果源是普通文件
		destAbsolutePath := destDirMatter.AbsolutePath() + "/" + name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//物理文件进行复制
		CopyFile(srcAbsolutePath, destAbsolutePath)

		//创建新文件的数据库信息。
		newMatter := &Matter{
			Puuid:    destDirMatter.Uuid,
			UserUuid: srcMatter.UserUuid,
			Username: srcMatter.Username,
			Dir:      srcMatter.Dir,
			Alien:    srcMatter.Alien,
			Name:     name,
			Md5:      "",
			Size:     srcMatter.Size,
			Privacy:  srcMatter.Privacy,
			Path:     destDirMatter.Path + "/" + name,
		}

		newMatter = this.matterDao.Create(newMatter)

	}

}

//将一个matter 重命名为 name
func (this *MatterService) Rename(matter *Matter, name string, user *User) {

	//验证参数。
	if name == "" {
		this.PanicBadRequest("name参数必填")
	}
	if m, _ := regexp.MatchString(`[<>|*?/\\]`, name); m {
		this.PanicBadRequest(`名称中不能包含以下特殊符号：< > | * ? / \`)
	}

	if len(name) > 200 {
		panic("name长度不能超过200")
	}

	if name == matter.Name {
		this.PanicBadRequest("新名称和旧名称一样，操作失败！")
	}

	//判断同级文件夹中是否有同名的文件
	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, matter.Puuid, matter.Dir, name)

	if count > 0 {
		this.PanicBadRequest("【" + name + "】已经存在了，请使用其他名称。")
	}

	if matter.Dir {
		//如果源是文件夹

		oldAbsolutePath := matter.AbsolutePath()
		absoluteDirPath := GetDirOfPath(oldAbsolutePath)
		relativeDirPath := GetDirOfPath(matter.Path)
		newAbsolutePath := absoluteDirPath + "/" + name

		//物理文件一口气移动
		err := os.Rename(oldAbsolutePath, newAbsolutePath)
		this.PanicError(err)

		//修改数据库中信息
		matter.Name = name
		matter.Path = relativeDirPath + "/" + name
		matter = this.matterDao.Save(matter)

		//调整该文件夹下文件的Path.
		matters := this.matterDao.List(matter.Uuid, matter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, matter)
		}

	} else {
		//如果源是普通文件

		oldAbsolutePath := matter.AbsolutePath()
		absoluteDirPath := GetDirOfPath(oldAbsolutePath)
		relativeDirPath := GetDirOfPath(matter.Path)
		newAbsolutePath := absoluteDirPath + "/" + name

		//物理文件进行移动
		err := os.Rename(oldAbsolutePath, newAbsolutePath)
		this.PanicError(err)

		//删除对应的缓存。
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		//修改数据库中信息
		matter.Name = name
		matter.Path = relativeDirPath + "/" + name
		matter = this.matterDao.Save(matter)

	}

	return
}
