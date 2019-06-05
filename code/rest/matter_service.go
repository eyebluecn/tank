package rest

import (
	"archive/zip"
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/download"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

/**
 * Methods start with Atomic has Lock. These method cannot be invoked by other Atomic Methods.
 */
//@Service
type MatterService struct {
	BaseBean
	matterDao         *MatterDao
	userDao           *UserDao
	userService       *UserService
	imageCacheDao     *ImageCacheDao
	imageCacheService *ImageCacheService
	preferenceService *PreferenceService
}

func (this *MatterService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.imageCacheService)
	if b, ok := b.(*ImageCacheService); ok {
		this.imageCacheService = b
	}

	b = core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

}

//Download. Support chunk download.
func (this *MatterService) DownloadFile(
	writer http.ResponseWriter,
	request *http.Request,
	filePath string,
	filename string,
	withContentDisposition bool) {

	download.DownloadFile(writer, request, filePath, filename, withContentDisposition)
}

//Download specified matters. matters must have the same puuid.
func (this *MatterService) DownloadZip(
	writer http.ResponseWriter,
	request *http.Request,
	matters []*Matter) {

	if matters == nil || len(matters) == 0 {
		panic(result.BadRequest("matters cannot be nil."))
	}
	userUuid := matters[0].UserUuid
	puuid := matters[0].Puuid

	for _, m := range matters {
		if m.UserUuid != userUuid {
			panic(result.BadRequest("userUuid not same"))
		} else if m.Puuid != puuid {
			panic(result.BadRequest("puuid not same"))
		}
	}

	preference := this.preferenceService.Fetch()

	//count the num of files will be downloaded.
	var count int64 = 0
	for _, matter := range matters {
		count = count + this.matterDao.CountByUserUuidAndPath(matter.UserUuid, matter.Path)
	}

	if preference.DownloadDirMaxNum >= 0 {
		if count > preference.DownloadDirMaxNum {
			panic(result.BadRequestI18n(request, i18n.MatterSelectNumExceedLimit, count, preference.DownloadDirMaxNum))
		}
	}

	//count the size of files will be downloaded.
	var sumSize int64 = 0
	for _, matter := range matters {
		sumSize = sumSize + this.matterDao.SumSizeByUserUuidAndPath(matter.UserUuid, matter.Path)
	}

	if preference.DownloadDirMaxSize >= 0 {
		if sumSize > preference.DownloadDirMaxSize {
			panic(result.BadRequestI18n(request, i18n.MatterSelectSizeExceedLimit, util.HumanFileSize(sumSize), util.HumanFileSize(preference.DownloadDirMaxSize)))
		}
	}

	//prepare the temp zip dir
	destZipDirPath := fmt.Sprintf("%s/%d", GetUserZipRootDir(matters[0].Username), time.Now().UnixNano()/1e6)
	util.MakeDirAll(destZipDirPath)

	destZipName := fmt.Sprintf("%s.zip", matters[0].Name)
	if len(matters) > 1 || !matters[0].Dir {
		destZipName = "archive.zip"
	}

	destZipPath := fmt.Sprintf("%s/%s", destZipDirPath, destZipName)

	this.zipMatters(request, matters, destZipPath)

	download.DownloadFile(writer, request, destZipPath, destZipName, true)

	//delete the temp zip file.
	err := os.Remove(destZipPath)
	if err != nil {
		this.logger.Error("error while deleting zip file %s", err.Error())
	}
	util.DeleteEmptyDir(destZipDirPath)

}

//zip matters.
func (this *MatterService) zipMatters(request *http.Request, matters []*Matter, destPath string) {

	if util.PathExists(destPath) {
		panic(result.BadRequest("%s exists", destPath))
	}

	//matters must have the same puuid.
	if matters == nil || len(matters) == 0 {
		panic(result.BadRequest("matters cannot be nil."))
	}
	userUuid := matters[0].UserUuid
	puuid := matters[0].Puuid
	baseDirPath := util.GetDirOfPath(matters[0].AbsolutePath()) + "/"

	for _, m := range matters {
		if m.UserUuid != userUuid {
			panic(result.BadRequest("userUuid not same"))
		} else if m.Puuid != puuid {
			panic(result.BadRequest("puuid not same"))
		}
	}

	//wrap children for every matter.
	for _, m := range matters {
		this.WrapChildrenDetail(request, m)
	}

	// create temp zip file.
	fileWriter, err := os.Create(destPath)
	this.PanicError(err)

	defer func() {
		err := fileWriter.Close()
		this.PanicError(err)
	}()

	zipWriter := zip.NewWriter(fileWriter)
	defer func() {
		err := zipWriter.Close()
		this.PanicError(err)
	}()

	//DFS algorithm
	var walkFunc func(matter *Matter)
	walkFunc = func(matter *Matter) {

		path := matter.AbsolutePath()

		fileInfo, err := os.Stat(path)
		this.PanicError(err)

		// Create file info.
		fileHeader, err := zip.FileInfoHeader(fileInfo)
		this.PanicError(err)

		// Trim the baseDirPath
		fileHeader.Name = strings.TrimPrefix(path, baseDirPath)

		// directory has prefix /
		if matter.Dir {
			fileHeader.Name += "/"
		}

		writer, err := zipWriter.CreateHeader(fileHeader)
		this.PanicError(err)

		// only regular file has things to write.
		if fileHeader.Mode().IsRegular() {

			fileToBeZip, err := os.Open(path)
			defer func() {
				err = fileToBeZip.Close()
				this.PanicError(err)
			}()
			this.PanicError(err)

			_, err = io.Copy(writer, fileToBeZip)
			this.PanicError(err)

		}

		//dfs.
		for _, m := range matter.Children {
			walkFunc(m)
		}
	}

	for _, m := range matters {
		walkFunc(m)
	}
}

//delete files.
func (this *MatterService) Delete(request *http.Request, matter *Matter, user *User) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	this.matterDao.Delete(matter)

	//re compute the size of Route.
	this.ComputeRouteSize(matter.Puuid, user)
}

//atomic delete files
func (this *MatterService) AtomicDelete(request *http.Request, matter *Matter, user *User) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	//lock
	this.userService.MatterLock(matter.UserUuid)
	defer this.userService.MatterUnlock(matter.UserUuid)

	this.Delete(request, matter, user)
}

//upload files.
func (this *MatterService) Upload(request *http.Request, file io.Reader, user *User, dirMatter *Matter, filename string, privacy bool) *Matter {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil."))
	}

	if len(filename) > MATTER_NAME_MAX_LENGTH {
		panic(result.BadRequestI18n(request, i18n.MatterNameLengthExceedLimit, len(filename), MATTER_NAME_MAX_LENGTH))
	}

	dirAbsolutePath := dirMatter.AbsolutePath()
	dirRelativePath := dirMatter.Path

	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, dirMatter.Uuid, false, filename)
	if count > 0 {
		panic(result.BadRequestI18n(request, i18n.MatterExist, filename))
	}

	fileAbsolutePath := dirAbsolutePath + "/" + filename
	fileRelativePath := dirRelativePath + "/" + filename

	util.MakeDirAll(dirAbsolutePath)

	//if exist, overwrite it.
	exist := util.PathExists(fileAbsolutePath)
	if exist {
		this.logger.Error("%s exits, overwrite it.", fileAbsolutePath)
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

	this.logger.Info("upload %s %v ", filename, util.HumanFileSize(fileSize))

	//check the size limit.
	if user.SizeLimit >= 0 {
		if fileSize > user.SizeLimit {
			//delete the file on disk.
			err = os.Remove(fileAbsolutePath)
			this.PanicError(err)

			panic(result.BadRequestI18n(request, i18n.MatterSizeExceedLimit, util.HumanFileSize(fileSize), util.HumanFileSize(user.SizeLimit)))
		}
	}

	//check total size.
	if user.TotalSizeLimit >= 0 {
		if user.TotalSize+fileSize > user.TotalSizeLimit {

			//delete the file on disk.
			err = os.Remove(fileAbsolutePath)
			this.PanicError(err)

			panic(result.BadRequestI18n(request, i18n.MatterSizeExceedTotalLimit, util.HumanFileSize(user.TotalSize), util.HumanFileSize(user.TotalSizeLimit)))
		}
	}

	//write to db.
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

	//compute the size of directory
	go core.RunWithRecovery(func() {
		this.ComputeRouteSize(dirMatter.Uuid, user)
	})

	return matter
}

// compute route size. It will compute upward until root directory
func (this *MatterService) ComputeRouteSize(matterUuid string, user *User) {

	//if to root directory, then update to user's info.
	if matterUuid == MATTER_ROOT {

		size := this.matterDao.SizeByPuuidAndUserUuid(MATTER_ROOT, user.Uuid)

		db := core.CONTEXT.GetDB().Model(&User{}).Where("uuid = ?", user.Uuid).Update("total_size", size)
		this.PanicError(db.Error)

		//update user total size info in cache.
		user.TotalSize = size

		return
	}

	matter := this.matterDao.CheckByUuid(matterUuid)

	//only compute dir
	if matter.Dir {
		//compute the total size.
		size := this.matterDao.SizeByPuuidAndUserUuid(matterUuid, user.Uuid)

		//when changed, we update
		if matter.Size != size {
			db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matterUuid).Update("size", size)
			this.PanicError(db.Error)
		}

	}

	//update parent recursively.
	this.ComputeRouteSize(matter.Puuid, user)
}

// compute all dir's size.
func (this *MatterService) ComputeAllDirSize(user *User) {

	this.logger.Info("Compute all dir's size for user %s %s", user.Uuid, user.Username)

	rootMatter := NewRootMatter(user)
	this.ComputeDirSize(rootMatter, user)
}

// compute a dir's size.
func (this *MatterService) ComputeDirSize(dirMatter *Matter, user *User) {

	this.logger.Info("Compute dir's size %s %s", dirMatter.Uuid, dirMatter.Name)

	//update sub dir first
	childrenDirMatters := this.matterDao.FindByUserUuidAndPuuidAndDirTrue(user.Uuid, dirMatter.Uuid)
	for _, childrenDirMatter := range childrenDirMatters {
		this.ComputeDirSize(childrenDirMatter, user)
	}

	//if to root directory, then update to user's info.
	if dirMatter.Uuid == MATTER_ROOT {

		size := this.matterDao.SizeByPuuidAndUserUuid(MATTER_ROOT, user.Uuid)

		db := core.CONTEXT.GetDB().Model(&User{}).Where("uuid = ?", user.Uuid).Update("total_size", size)
		this.PanicError(db.Error)

		//update user total size info in cache.
		user.TotalSize = size
	} else {

		//compute self.
		size := this.matterDao.SizeByPuuidAndUserUuid(dirMatter.Uuid, user.Uuid)

		//when changed, we update
		if dirMatter.Size != size {
			db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", dirMatter.Uuid).Update("size", size)
			this.PanicError(db.Error)
		}
	}

}

//inner create directory.
func (this *MatterService) createDirectory(request *http.Request, dirMatter *Matter, name string, user *User) *Matter {

	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil"))
	}

	if !dirMatter.Dir {
		panic(result.BadRequest("dirMatter must be directory"))
	}

	if dirMatter.UserUuid != user.Uuid {

		panic(result.BadRequest("file's user not the same"))
	}

	name = strings.TrimSpace(name)
	if name == "" {
		panic(result.BadRequest("name cannot be blank"))
	}

	if len(name) > MATTER_NAME_MAX_LENGTH {

		panic(result.BadRequestI18n(request, i18n.MatterNameLengthExceedLimit, len(name), MATTER_NAME_MAX_LENGTH))

	}

	if m, _ := regexp.MatchString(MATTER_NAME_PATTERN, name); m {
		panic(result.BadRequestI18n(request, i18n.MatterNameContainSpecialChars))
	}

	//if exist. return.
	matter := this.matterDao.FindByUserUuidAndPuuidAndDirAndName(user.Uuid, dirMatter.Uuid, true, name)
	if matter != nil {
		return matter
	}

	parts := strings.Split(dirMatter.Path, "/")

	if len(parts) > MATTER_NAME_MAX_DEPTH {
		panic(result.BadRequestI18n(request, i18n.MatterDepthExceedLimit, len(parts), MATTER_NAME_MAX_DEPTH))
	}

	absolutePath := GetUserMatterRootDir(user.Username) + dirMatter.Path + "/" + name

	relativePath := dirMatter.Path + "/" + name

	//crate directory on disk.
	dirPath := util.MakeDirAll(absolutePath)
	this.logger.Info("Create Directory: %s", dirPath)

	//create in db
	matter = &Matter{
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

func (this *MatterService) AtomicCreateDirectory(request *http.Request, dirMatter *Matter, name string, user *User) *Matter {

	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	matter := this.createDirectory(request, dirMatter, name, user)

	return matter
}

//copy or move may overwrite.
func (this *MatterService) handleOverwrite(request *http.Request, user *User, destinationPath string, overwrite bool) {

	destMatter := this.matterDao.findByUserUuidAndPath(user.Uuid, destinationPath)
	if destMatter != nil {
		//if exist
		if overwrite {
			//delete.
			this.Delete(request, destMatter, user)
		} else {
			panic(result.BadRequestI18n(request, i18n.MatterExist, destMatter.Path))
		}
	}

}

//move srcMatter to destMatter. invoker must handled the overwrite and lock.
func (this *MatterService) move(request *http.Request, srcMatter *Matter, destDirMatter *Matter, user *User) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	if !destDirMatter.Dir {
		panic(result.BadRequestI18n(request, i18n.MatterDestinationMustDirectory))
	}

	srcPuuid := srcMatter.Puuid
	destDirUuid := destDirMatter.Uuid

	if srcMatter.Dir {

		//if src is dir.
		destAbsolutePath := destDirMatter.AbsolutePath() + "/" + srcMatter.Name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//move src to dest on disk.
		err := os.Rename(srcAbsolutePath, destAbsolutePath)
		this.PanicError(err)

		//change info on db.
		srcMatter.Puuid = destDirMatter.Uuid
		srcMatter.Path = destDirMatter.Path + "/" + srcMatter.Name
		srcMatter = this.matterDao.Save(srcMatter)

		//reCompute the path.
		matters := this.matterDao.FindByPuuidAndUserUuid(srcMatter.Uuid, srcMatter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, srcMatter)
		}

	} else {

		//if src is NOT dir.
		destAbsolutePath := destDirMatter.AbsolutePath() + "/" + srcMatter.Name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//move src to dest on disk.
		err := os.Rename(srcAbsolutePath, destAbsolutePath)
		this.PanicError(err)

		//delete caches.
		this.imageCacheDao.DeleteByMatterUuid(srcMatter.Uuid)

		//change info in db.
		srcMatter.Puuid = destDirMatter.Uuid
		srcMatter.Path = destDirMatter.Path + "/" + srcMatter.Name
		srcMatter = this.matterDao.Save(srcMatter)

	}

	//reCompute the size of src and dest.
	this.ComputeRouteSize(srcPuuid, user)
	this.ComputeRouteSize(destDirUuid, user)

}

//move srcMatter to destMatter(must be dir)
func (this *MatterService) AtomicMove(request *http.Request, srcMatter *Matter, destDirMatter *Matter, overwrite bool, user *User) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	this.userService.MatterLock(srcMatter.UserUuid)
	defer this.userService.MatterUnlock(srcMatter.UserUuid)

	if destDirMatter == nil {
		panic(result.BadRequest("destDirMatter cannot be nil."))
	}
	if !destDirMatter.Dir {
		panic(result.BadRequestI18n(request, i18n.MatterDestinationMustDirectory))
	}

	//neither move to itself, nor move to its children.
	destDirMatter = this.WrapParentDetail(request, destDirMatter)
	tmpMatter := destDirMatter
	for tmpMatter != nil {
		if srcMatter.Uuid == tmpMatter.Uuid {
			panic(result.BadRequestI18n(request, i18n.MatterMoveRecursive))
		}
		tmpMatter = tmpMatter.Parent
	}

	//handle the overwrite
	destinationPath := destDirMatter.Path + "/" + srcMatter.Name
	this.handleOverwrite(request, user, destinationPath, overwrite)

	//do the move operation.
	this.move(request, srcMatter, destDirMatter, user)
}

//move srcMatters to destMatter(must be dir)
func (this *MatterService) AtomicMoveBatch(request *http.Request, srcMatters []*Matter, destDirMatter *Matter, user *User) {

	if destDirMatter == nil {
		panic(result.BadRequest("destDirMatter cannot be nil."))
	}

	this.userService.MatterLock(destDirMatter.UserUuid)
	defer this.userService.MatterUnlock(destDirMatter.UserUuid)

	if srcMatters == nil {
		panic(result.BadRequest("srcMatters cannot be nil."))
	}

	if !destDirMatter.Dir {
		panic(result.BadRequestI18n(request, i18n.MatterDestinationMustDirectory))
	}

	//neither move to itself, nor move to its children.
	destDirMatter = this.WrapParentDetail(request, destDirMatter)
	for _, srcMatter := range srcMatters {

		tmpMatter := destDirMatter
		for tmpMatter != nil {
			if srcMatter.Uuid == tmpMatter.Uuid {
				panic(result.BadRequestI18n(request, i18n.MatterMoveRecursive))
			}
			tmpMatter = tmpMatter.Parent
		}
	}

	for _, srcMatter := range srcMatters {
		this.move(request, srcMatter, destDirMatter, user)
	}

}

//copy srcMatter to destMatter. invoker must handled the overwrite and lock.
func (this *MatterService) copy(request *http.Request, srcMatter *Matter, destDirMatter *Matter, name string) {

	if srcMatter.Dir {

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

		//copy children
		matters := this.matterDao.FindByPuuidAndUserUuid(srcMatter.Uuid, srcMatter.UserUuid, nil)
		for _, m := range matters {
			this.copy(request, m, newMatter, m.Name)
		}

	} else {

		destAbsolutePath := destDirMatter.AbsolutePath() + "/" + name
		srcAbsolutePath := srcMatter.AbsolutePath()

		//copy file on disk.
		util.CopyFile(srcAbsolutePath, destAbsolutePath)

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

//copy srcMatter to destMatter.
func (this *MatterService) AtomicCopy(request *http.Request, srcMatter *Matter, destDirMatter *Matter, name string, overwrite bool, user *User) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	this.userService.MatterLock(srcMatter.UserUuid)
	defer this.userService.MatterUnlock(srcMatter.UserUuid)

	if !destDirMatter.Dir {
		panic(result.BadRequestI18n(request, i18n.MatterDestinationMustDirectory))
	}

	destinationPath := destDirMatter.Path + "/" + name
	this.handleOverwrite(request, user, destinationPath, overwrite)

	this.copy(request, srcMatter, destDirMatter, name)
}

//rename matter to name
func (this *MatterService) AtomicRename(request *http.Request, matter *Matter, name string, user *User) {

	if user == nil {
		panic(result.BadRequest("user cannot be nil"))
	}

	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	name = CheckMatterName(request, name)

	if name == matter.Name {
		panic(result.BadRequestI18n(request, i18n.MatterNameNoChange))
	}

	//判断同级文件夹中是否有同名的文件
	count := this.matterDao.CountByUserUuidAndPuuidAndDirAndName(user.Uuid, matter.Puuid, matter.Dir, name)

	if count > 0 {

		panic(result.BadRequestI18n(request, i18n.MatterExist, name))
	}

	if matter.Dir {
		//如果源是文件夹

		oldAbsolutePath := matter.AbsolutePath()
		absoluteDirPath := util.GetDirOfPath(oldAbsolutePath)
		relativeDirPath := util.GetDirOfPath(matter.Path)
		newAbsolutePath := absoluteDirPath + "/" + name

		//物理文件一口气移动
		err := os.Rename(oldAbsolutePath, newAbsolutePath)
		this.PanicError(err)

		//修改数据库中信息
		matter.Name = name
		matter.Path = relativeDirPath + "/" + name
		matter = this.matterDao.Save(matter)

		//调整该文件夹下文件的Path.
		matters := this.matterDao.FindByPuuidAndUserUuid(matter.Uuid, matter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, matter)
		}

	} else {
		//如果源是普通文件

		oldAbsolutePath := matter.AbsolutePath()
		absoluteDirPath := util.GetDirOfPath(oldAbsolutePath)
		relativeDirPath := util.GetDirOfPath(matter.Path)
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

//将本地文件映射到蓝眼云盘中去。
func (this *MatterService) AtomicMirror(request *http.Request, srcPath string, destPath string, overwrite bool, user *User) {

	if user == nil {
		panic(result.BadRequest("user cannot be nil"))
	}

	//操作锁
	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	//验证参数。
	if destPath == "" {
		panic(result.BadRequest("dest cannot be null"))
	}

	destDirMatter := this.CreateDirectories(request, user, destPath)

	this.mirror(request, srcPath, destDirMatter, overwrite, user)
}

//将本地文件/文件夹映射到蓝眼云盘中去。
func (this *MatterService) mirror(request *http.Request, srcPath string, destDirMatter *Matter, overwrite bool, user *User) {

	if user == nil {
		panic(result.BadRequest("user cannot be nil"))
	}

	fileStat, err := os.Stat(srcPath)
	if err != nil {

		if os.IsNotExist(err) {
			panic(result.BadRequest("srcPath %s not exist", srcPath))
		} else {

			panic(result.BadRequest("srcPath err %s %s", srcPath, err.Error()))
		}

	}

	this.logger.Info("mirror srcPath = %s destPath = %s", srcPath, destDirMatter.Path)

	if fileStat.IsDir() {

		//判断当前文件夹下，文件是否已经存在了。
		srcDirMatter := this.matterDao.FindByUserUuidAndPuuidAndDirAndName(user.Uuid, destDirMatter.Uuid, true, fileStat.Name())

		if srcDirMatter == nil {
			srcDirMatter = this.createDirectory(request, destDirMatter, fileStat.Name(), user)
		}

		fileInfos, err := ioutil.ReadDir(srcPath)
		this.PanicError(err)

		//递归处理本文件夹下的文件或文件夹
		for _, fileInfo := range fileInfos {

			path := fmt.Sprintf("%s/%s", srcPath, fileInfo.Name())
			this.mirror(request, path, srcDirMatter, overwrite, user)
		}

	} else {

		//判断当前文件夹下，文件是否已经存在了。
		matter := this.matterDao.FindByUserUuidAndPuuidAndDirAndName(user.Uuid, destDirMatter.Uuid, false, fileStat.Name())
		if matter != nil {
			//如果是覆盖，那么删除之前的文件
			if overwrite {
				this.Delete(request, matter, user)
			} else {
				//直接完成。
				return
			}
		}

		//准备直接从本地上传了。
		file, err := os.Open(srcPath)
		this.PanicError(err)
		defer func() {
			err := file.Close()
			this.PanicError(err)
		}()

		this.Upload(request, file, user, destDirMatter, fileStat.Name(), true)

	}

}

//根据一个文件夹路径，依次创建，找到最后一个文件夹的matter，如果中途出错，返回err. 如果存在了那就直接返回即可。
func (this *MatterService) CreateDirectories(request *http.Request, user *User, dirPath string) *Matter {

	dirPath = path.Clean(dirPath)

	if dirPath == "/" {
		return NewRootMatter(user)
	}

	//ignore the last slash.
	dirPath = strings.TrimSuffix(dirPath, "/")

	folders := strings.Split(dirPath, "/")

	if len(folders) > MATTER_NAME_MAX_DEPTH {
		panic(result.BadRequestI18n(request, i18n.MatterDepthExceedLimit, len(folders), MATTER_NAME_MAX_DEPTH))
	}

	//validate every matter name.
	for k, name := range folders {
		//first element is ""
		if k != 0 {
			CheckMatterName(request, name)
		}
	}

	var dirMatter *Matter
	for k, name := range folders {

		//ignore the first element.
		if k == 0 {
			dirMatter = NewRootMatter(user)
			continue
		}

		dirMatter = this.createDirectory(request, dirMatter, name, user)
	}

	return dirMatter
}

//wrap a matter. put its parent.
func (this *MatterService) WrapParentDetail(request *http.Request, matter *Matter) *Matter {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil."))
	}

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

//wrap a matter ,put its children
func (this *MatterService) WrapChildrenDetail(request *http.Request, matter *Matter) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil."))
	}

	if matter.Dir {

		children := this.matterDao.FindByPuuidAndUserUuid(matter.Uuid, matter.UserUuid, nil)
		matter.Children = children

		for _, child := range matter.Children {
			this.WrapChildrenDetail(request, child)
		}
	}

}

//fetch a matter's detail with parent info.
func (this *MatterService) Detail(request *http.Request, uuid string) *Matter {
	matter := this.matterDao.CheckByUuid(uuid)
	return this.WrapParentDetail(request, matter)
}

//crawl a url to dirMatter
func (this *MatterService) AtomicCrawl(request *http.Request, url string, filename string, user *User, dirMatter *Matter, privacy bool) *Matter {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		panic(`url must start with http:// or https://`)
	}

	if filename == "" {
		panic(result.BadRequest("filename cannot be null."))
	}

	//download from url.
	resp, err := http.Get(url)
	this.PanicError(err)

	return this.Upload(request, resp.Body, user, dirMatter, filename, privacy)
}

//adjust a matter's path.
func (this *MatterService) adjustPath(matter *Matter, parentMatter *Matter) {

	if matter.Dir {

		matter.Path = parentMatter.Path + "/" + matter.Name
		matter = this.matterDao.Save(matter)

		//adjust children.
		matters := this.matterDao.FindByPuuidAndUserUuid(matter.Uuid, matter.UserUuid, nil)
		for _, m := range matters {
			this.adjustPath(m, matter)
		}

	} else {
		//delete caches.
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		matter.Path = parentMatter.Path + "/" + matter.Name
		matter = this.matterDao.Save(matter)
	}

}
