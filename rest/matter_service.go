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
func (this *MatterService) Upload(file io.Reader, user *User, dirMatter *Matter, filename string, privacy bool) *Matter {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	//操作锁
	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	//验证dirMatter
	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil."))
	}

	//文件名不能太长。
	if len(filename) > MATTER_NAME_MAX_LENGTH {
		panic(result.BadRequest("文件名不能超过%s", MATTER_NAME_MAX_LENGTH))
	}

	//文件夹路径
	dirAbsolutePath := dirMatter.AbsolutePath()
	dirRelativePath := dirMatter.Path

	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, dirMatter.Uuid, false, filename)
	if count > 0 {
		panic(result.BadRequest("该目录下%s已经存在了", filename))
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

	destFile, err := os.OpenFile(fileAbsolutePath, os.O_WRONLY|os.O_CREATE, 0777)
	this.PanicError(err)

	defer func() {
		err := destFile.Close()
		this.PanicError(err)
	}()

	fileSize, err := io.Copy(destFile, file)
	this.PanicError(err)

	this.logger.Info("上传文件 %s 大小为 %v ", filename, HumanFileSize(fileSize))

	//判断用户自身上传大小的限制。
	if user.SizeLimit >= 0 {
		if fileSize > user.SizeLimit {
			//删除上传过来的内容
			err = os.Remove(fileAbsolutePath)
			this.PanicError(err)

			panic(result.BadRequest("文件大小超出限制 %s > %s ", HumanFileSize(user.SizeLimit), HumanFileSize(fileSize)))
		}
	}

	//将文件信息存入数据库中。
	matter := &Matter{
		Puuid:    dirMatter.Uuid,
		UserUuid: user.Uuid,
		Username: user.Username,
		Dir:      false,
		Name:     filename,
		Md5:      "",
		Size:     fileSize,
		Privacy:  privacy,
		Path:     fileRelativePath,
	}
	matter = this.matterDao.Create(matter)

	return matter
}

//在dirMatter中创建文件夹 返回刚刚创建的这个文件夹
func (this *MatterService) CreateDirectory(dirMatter *Matter, name string, user *User) *Matter {

	//操作锁
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

//处理 移动和复制时可能存在的覆盖问题。
func (this *MatterService) handleOverwrite(userUuid string, destinationPath string, overwrite bool) {

	//目标matter。因为有可能已经存在了
	destMatter := this.matterDao.findByUserUuidAndPath(userUuid, destinationPath)
	//如果目标matter存在了。
	if destMatter != nil {
		//如果目标matter还存在了。
		if overwrite {
			//要求覆盖。那么删除。
			this.matterDao.Delete(destMatter)
		} else {
			this.PanicBadRequest("%s已经存在，操作失败！", destMatter.Path)
		}
	}

}

//将一个srcMatter放置到另一个destMatter(必须为文件夹)下 不关注 overwrite 和 lock.
func (this *MatterService) innerMove(srcMatter *Matter, destDirMatter *Matter) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	if !destDirMatter.Dir {
		this.PanicBadRequest("目标必须为文件夹")
	}

	if srcMatter.Dir {
		//如果源是文件夹
		destAbsolutePath := destDirMatter.AbsolutePath() + "/" + srcMatter.Name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//物理文件一口气移动
		err := os.Rename(srcAbsolutePath, destAbsolutePath)
		this.PanicError(err)

		//修改数据库中信息
		srcMatter.Puuid = destDirMatter.Uuid
		srcMatter.Path = destDirMatter.Path + "/" + srcMatter.Name
		srcMatter = this.matterDao.Save(srcMatter)

		//调整该文件夹下文件的Path.
		matters := this.matterDao.List(srcMatter.Uuid, srcMatter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, srcMatter)
		}

	} else {
		//如果源是普通文件

		destAbsolutePath := destDirMatter.AbsolutePath() + "/" + srcMatter.Name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//物理文件进行移动
		err := os.Rename(srcAbsolutePath, destAbsolutePath)
		this.PanicError(err)

		//删除对应的缓存。
		this.imageCacheDao.DeleteByMatterUuid(srcMatter.Uuid)

		//修改数据库中信息
		srcMatter.Puuid = destDirMatter.Uuid
		srcMatter.Path = destDirMatter.Path + "/" + srcMatter.Name
		srcMatter = this.matterDao.Save(srcMatter)

	}

	return
}

//将一个srcMatter放置到另一个destMatter(必须为文件夹)下
func (this *MatterService) Move(srcMatter *Matter, destDirMatter *Matter, overwrite bool) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	//操作锁
	this.userService.MatterLock(srcMatter.UserUuid)
	defer this.userService.MatterUnlock(srcMatter.UserUuid)

	if !destDirMatter.Dir {
		this.PanicBadRequest("目标必须为文件夹")
	}


	//文件夹不能把自己移入到自己中，也不可以移入到自己的子文件夹下。
	destDirMatter = this.WrapDetail(destDirMatter)
	tmpMatter := destDirMatter
	for tmpMatter != nil {
		if srcMatter.Uuid == tmpMatter.Uuid {
			panic("文件夹不能把自己移入到自己中，也不可以移入到自己的子文件夹下。")
		}
		tmpMatter = tmpMatter.Parent
	}

	//处理覆盖的问题
	destinationPath := destDirMatter.Path + "/" + srcMatter.Name
	this.handleOverwrite(srcMatter.UserUuid, destinationPath, overwrite)

	//做move操作。
	this.innerMove(srcMatter, destDirMatter)
}

//将一个srcMatter放置到另一个destMatter(必须为文件夹)下
func (this *MatterService) MoveBatch(srcMatters []*Matter, destDirMatter *Matter) {

	if destDirMatter == nil {
		panic(result.BadRequest("destDirMatter cannot be nil."))
	}

	//操作锁
	this.userService.MatterLock(destDirMatter.UserUuid)
	defer this.userService.MatterUnlock(destDirMatter.UserUuid)

	if srcMatters == nil {
		panic(result.BadRequest("srcMatters cannot be nil."))
	}

	if !destDirMatter.Dir {
		this.PanicBadRequest("目标必须为文件夹")
	}

	//文件夹不能把自己移入到自己中，也不可以移入到自己的子文件夹下。
	destDirMatter = this.WrapDetail(destDirMatter)
	for _, srcMatter := range srcMatters {

		tmpMatter := destDirMatter
		for tmpMatter != nil {
			if srcMatter.Uuid == tmpMatter.Uuid {
				panic("文件夹不能把自己移入到自己中，也不可以移入到自己的子文件夹下。")
			}
			tmpMatter = tmpMatter.Parent
		}
	}

	for _, srcMatter := range srcMatters {
		this.innerMove(srcMatter, destDirMatter)
	}

}

//内部移动一个文件(提供给Copy调用)，无需关心overwrite问题。
func (this *MatterService) innerCopy(srcMatter *Matter, destDirMatter *Matter, name string) {

	if srcMatter.Dir {

		//如果源是文件夹

		//在目标地址创建新文件夹。
		newMatter := &Matter{
			Puuid:    destDirMatter.Uuid,
			UserUuid: srcMatter.UserUuid,
			Username: srcMatter.Username,
			Dir:      srcMatter.Dir,
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
			this.innerCopy(m, newMatter, m.Name)
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
			Name:     name,
			Md5:      "",
			Size:     srcMatter.Size,
			Privacy:  srcMatter.Privacy,
			Path:     destDirMatter.Path + "/" + name,
		}
		newMatter = this.matterDao.Create(newMatter)

	}
}

//将一个srcMatter复制到另一个destMatter(必须为文件夹)下，名字叫做name
func (this *MatterService) Copy(srcMatter *Matter, destDirMatter *Matter, name string, overwrite bool) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	//操作锁
	this.userService.MatterLock(srcMatter.UserUuid)
	defer this.userService.MatterUnlock(srcMatter.UserUuid)

	if !destDirMatter.Dir {
		this.PanicBadRequest("目标必须为文件夹")
	}

	destinationPath := destDirMatter.Path + "/" + name
	this.handleOverwrite(srcMatter.UserUuid, destinationPath, overwrite)

	this.innerCopy(srcMatter, destDirMatter, name)
}

//将一个matter 重命名为 name
func (this *MatterService) Rename(matter *Matter, name string, user *User) {

	if user == nil {
		this.PanicBadRequest("user cannot be nil")
	}

	//操作锁
	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

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

//包装某个matter的详情。会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *MatterService) WrapDetail(matter *Matter) *Matter {

	if matter == nil {
		this.PanicBadRequest("matter cannot be nil.")
	}

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

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *MatterService) Detail(uuid string) *Matter {
	matter := this.matterDao.CheckByUuid(uuid)
	return this.WrapDetail(matter)
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
	size, err := download.HttpDownloadFile(absolutePath, url)
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
