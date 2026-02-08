package rest

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/download"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
)

/**
 * Methods start with Atomic has Lock. These method cannot be invoked by other Atomic Methods.
 */
//@Service
type MatterService struct {
	BaseBean
	matterDao         *MatterDao
	spaceDao          *SpaceDao
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

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
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

// get the page of matters.
func (this *MatterService) Page(
	request *http.Request,
	page int,
	pageSize int,
	orderCreateTime string,
	orderUpdateTime string,
	orderDeleteTime string,
	orderSort string,
	orderTimes string,
	orderDir string,
	orderSize string,
	orderName string,
	puuid string,
	name string,
	dir string,
	deleted string,
	extensions []string,
	spaceUuid string,
) *Pager {

	sortArray := []builder.OrderPair{
		{
			Key:   "dir",
			Value: orderDir,
		},
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
		{
			Key:   "update_time",
			Value: orderUpdateTime,
		},
		{
			Key:   "delete_time",
			Value: orderDeleteTime,
		},
		{
			Key:   "sort",
			Value: orderSort,
		},
		{
			Key:   "size",
			Value: orderSize,
		},
		{
			Key:   "name",
			Value: orderName,
		},
		{
			Key:   "times",
			Value: orderTimes,
		},
	}

	pager := this.matterDao.Page(page, pageSize, puuid, "", spaceUuid, name, dir, deleted, extensions, sortArray)

	return pager
}

// search files by dfs.
func (this *MatterService) DfsSearch(
	request *http.Request,
	limit int,
	puuid string,
	keyword string,
	spaceUuid string,
	deleted bool,
) []*Matter {

	deletedStr := FALSE
	if deleted {
		deletedStr = FALSE
	}

	//find all matters including dir and files.
	_, matters := this.matterDao.PlainPage(0, limit, puuid, "", spaceUuid, keyword, "", deletedStr, nil, nil, nil)
	if len(matters) >= limit {
		return matters
	}

	var resultList = make([]*Matter, 0)
	for _, matter := range matters {
		if !matter.Dir {
			resultList = append(resultList, matter)
		}
	}

	//dfs from puuid.
	_, dirMatters := this.matterDao.PlainPage(0, 1000, puuid, "", spaceUuid, "", TRUE, "", nil, nil, nil)
	for _, dirMatter := range dirMatters {

		//add dir if match.
		if strings.Contains(dirMatter.Name, keyword) {
			if deleted == dirMatter.Deleted {
				resultList = append(resultList, dirMatter)
			}
		}

		remainLimit := limit - len(resultList)
		subMatters := this.DfsSearch(request, remainLimit, dirMatter.Uuid, keyword, spaceUuid, deleted)
		for _, subMatter := range subMatters {
			resultList = append(resultList, subMatter)
		}

		//enough then return
		if len(resultList) >= limit {
			return resultList
		}

	}

	//not enough ,return.
	return resultList
}

// Download. Support chunk download.
func (this *MatterService) DownloadFile(
	writer http.ResponseWriter,
	request *http.Request,
	filePath string,
	filename string,
	withContentDisposition bool) {

	download.DownloadFile(writer, request, filePath, filename, withContentDisposition)
}

// Download specified matters. matters must have the same puuid.
func (this *MatterService) DownloadZip(
	writer http.ResponseWriter,
	request *http.Request,
	matters []*Matter) {

	if matters == nil || len(matters) == 0 {
		panic(result.BadRequest("matters cannot be nil."))
	}
	spaceUuid := matters[0].SpaceUuid
	puuid := matters[0].Puuid

	for _, m := range matters {
		if m.SpaceUuid != spaceUuid {
			panic(result.BadRequest("spaceUuid not same"))
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
	destZipDirPath := fmt.Sprintf("%s/%d", GetSpaceZipRootDir(matters[0].SpaceName), time.Now().UnixNano()/1e6)
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

// zip matters.
func (this *MatterService) zipMatters(request *http.Request, matters []*Matter, destPath string) {

	if util.PathExists(destPath) {
		panic(result.BadRequest("%s exists", destPath))
	}

	//matters must have the same puuid.
	if matters == nil || len(matters) == 0 {
		panic(result.BadRequest("matters cannot be nil."))
	}
	spaceUuid := matters[0].SpaceUuid
	puuid := matters[0].Puuid
	baseDirPath := util.GetDirOfPath(matters[0].AbsolutePath()) + "/"

	for _, m := range matters {
		if m.SpaceUuid != spaceUuid {
			panic(result.BadRequest("spaceUuid not same"))
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

// delete files.
func (this *MatterService) Delete(request *http.Request, matter *Matter, user *User, space *Space) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	this.matterDao.Delete(matter)

	//re compute the size of Route.
	this.ComputeRouteSize(matter.Puuid, user, space)
}

// soft delete files.
func (this *MatterService) SoftDelete(request *http.Request, matter *Matter, user *User) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	if matter.Deleted {
		panic(result.BadRequest("matter has been deleted"))
	}

	this.matterDao.SoftDelete(matter)
	//no need to recompute size.
}

// recovery delete files.
func (this *MatterService) Recovery(request *http.Request, matter *Matter, user *User) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	if !matter.Deleted {
		panic(result.BadRequest("matter has not been deleted"))
	}

	this.matterDao.Recovery(matter)
	//no need to recompute size.
}

// atomic delete files
func (this *MatterService) AtomicDelete(request *http.Request, matter *Matter, user *User, space *Space) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	//lock
	this.userService.MatterLock(matter.UserUuid)
	defer this.userService.MatterUnlock(matter.UserUuid)

	this.Delete(request, matter, user, space)
}

// atomic soft delete files
func (this *MatterService) AtomicSoftDelete(request *http.Request, matter *Matter, user *User, space *Space) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	if matter.Deleted {
		panic(result.BadRequest("matter has been deleted"))
	}

	//lock
	this.userService.MatterLock(matter.UserUuid)
	defer this.userService.MatterUnlock(matter.UserUuid)

	//if disabled the recycle feature. then we hard delete.
	preference := this.preferenceService.Fetch()
	if preference.DeletedKeepDays == 0 {
		this.Delete(request, matter, user, space)
	} else {
		this.SoftDelete(request, matter, user)
	}

}

// atomic recovery delete files
func (this *MatterService) AtomicRecovery(request *http.Request, matter *Matter, user *User) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil"))
	}

	if !matter.Deleted {
		panic(result.BadRequest("matter has not been deleted"))
	}

	//lock
	this.userService.MatterLock(matter.UserUuid)
	defer this.userService.MatterUnlock(matter.UserUuid)

	this.Recovery(request, matter, user)
}

// upload files.
func (this *MatterService) Upload(request *http.Request, file io.Reader, fileHeader *multipart.FileHeader, user *User, space *Space, dirMatter *Matter, filename string, privacy bool) *Matter {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil."))
	}

	if dirMatter.Deleted {
		panic(result.BadRequest("Dir has been deleted. Cannot upload under it."))
	}

	CheckMatterName(request, filename)

	//if fileHeader.Size not nill . check size in advance.
	if fileHeader != nil {
		//check the size limit.
		if space.SizeLimit >= 0 {
			if fileHeader.Size > space.SizeLimit {
				panic(result.BadRequestI18n(request, i18n.MatterSizeExceedLimit, util.HumanFileSize(fileHeader.Size), util.HumanFileSize(space.SizeLimit)))
			}
		}

		//check total size.
		if space.TotalSizeLimit >= 0 {
			if space.TotalSize+fileHeader.Size > space.TotalSizeLimit {
				panic(result.BadRequestI18n(request, i18n.MatterSizeExceedTotalLimit, util.HumanFileSize(space.TotalSize), util.HumanFileSize(space.TotalSizeLimit)))
			}
		}
	}

	dirAbsolutePath := dirMatter.AbsolutePath()

	dbMatter := this.matterDao.FindBySpaceUuidAndPuuidAndDirAndName(space.Uuid, dirMatter.Uuid, false, filename)
	if dbMatter != nil {
		if dbMatter.Deleted {
			panic(result.BadRequestI18n(request, i18n.MatterRecycleBinExist, filename))
		} else {
			panic(result.BadRequestI18n(request, i18n.MatterExist, filename))
		}

	}

	fileAbsolutePath := dirAbsolutePath + "/" + filename

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
	closeDestFile := func() {
		err := destFile.Close()
		this.PanicError(err)
	}

	fileSize, err := io.Copy(destFile, file)
	this.PanicError(err)

	this.logger.Info("upload %s %v ", filename, util.HumanFileSize(fileSize))

	if fileHeader == nil {
		//check the size limit.
		if space.SizeLimit >= 0 {
			if fileSize > space.SizeLimit {
				closeDestFile()

				//delete the file on disk.
				err = os.Remove(fileAbsolutePath)
				this.PanicError(err)

				panic(result.BadRequestI18n(request, i18n.MatterSizeExceedLimit, util.HumanFileSize(fileSize), util.HumanFileSize(space.SizeLimit)))
			}
		}

		//check total size.
		if space.TotalSizeLimit >= 0 {
			if space.TotalSize+fileSize > space.TotalSizeLimit {
				closeDestFile()

				//delete the file on disk.
				err = os.Remove(fileAbsolutePath)
				this.PanicError(err)

				panic(result.BadRequestI18n(request, i18n.MatterSizeExceedTotalLimit, util.HumanFileSize(space.TotalSize), util.HumanFileSize(space.TotalSizeLimit)))
			}
		}
	}

	closeDestFile()

	matter := this.createNonDirMatter(dirMatter, filename, fileSize, privacy, user, space)

	return matter
}

// create a non dir matter.
func (this *MatterService) createNonDirMatter(dirMatter *Matter, filename string, fileSize int64, privacy bool, user *User, space *Space) *Matter {
	dirRelativePath := dirMatter.Path
	fileRelativePath := dirRelativePath + "/" + filename

	//write to db.
	matter := &Matter{
		Puuid:     dirMatter.Uuid,
		UserUuid:  user.Uuid,
		SpaceName: space.Name,
		SpaceUuid: space.Uuid,
		Dir:       false,
		Name:      filename,
		Md5:       "",
		Size:      fileSize,
		Privacy:   privacy,
		Path:      fileRelativePath,
		Prop:      EMPTY_JSON_MAP,
		VisitTime: time.Now(),
	}
	matter = this.matterDao.Create(matter)

	//compute the size of directory
	go core.RunWithRecovery(func() {
		this.ComputeRouteSize(dirMatter.Uuid, user, space)
	})

	return matter
}

// create a non dir matter.
func (this *MatterService) updateNonDirMatter(matter *Matter, fileSize int64, user *User, space *Space) *Matter {

	matter.Size = fileSize

	matter = this.matterDao.Save(matter)

	//compute the size of directory
	go core.RunWithRecovery(func() {
		this.ComputeRouteSize(matter.Puuid, user, space)
	})

	return matter
}

// compute route size. It will compute upward until root directory
func (this *MatterService) ComputeRouteSize(matterUuid string, user *User, space *Space) {

	//if to root directory, then update to user's info.
	if matterUuid == MATTER_ROOT {

		size := this.matterDao.SizeByPuuidAndSpaceUuid(MATTER_ROOT, space.Uuid)

		this.spaceDao.UpdateTotalSize(space.Uuid, size)

		//update user total size info in cache.
		space.TotalSize = size

		return
	}

	matter := this.matterDao.CheckByUuid(matterUuid)

	//only compute dir
	if matter.Dir {
		//compute the total size.
		size := this.matterDao.SizeByPuuidAndSpaceUuid(matterUuid, space.Uuid)

		//when changed, we update
		if matter.Size != size {
			this.matterDao.UpdateSize(matterUuid, size)

			matter.Size = size
		}

	}

	//update parent recursively.
	this.ComputeRouteSize(matter.Puuid, user, space)
}

// compute all dir's size.
func (this *MatterService) ComputeAllDirSize(user *User, space *Space) {

	this.logger.Info("Compute all dir's size for user %s %s", user.Uuid, user.Username)

	rootMatter := NewRootMatter(space)
	this.ComputeDirSize(rootMatter, user, space)
}

// compute a dir's size.
func (this *MatterService) ComputeDirSize(dirMatter *Matter, user *User, space *Space) {

	this.logger.Info("Compute dir's size %s %s", dirMatter.Uuid, dirMatter.Name)

	//update sub dir first
	childrenDirMatters := this.matterDao.FindByUserUuidAndPuuidAndDirTrue(user.Uuid, dirMatter.Uuid)
	for _, childrenDirMatter := range childrenDirMatters {
		this.ComputeDirSize(childrenDirMatter, user, space)
	}

	//if to root directory, then update to user's info.
	if dirMatter.Uuid == MATTER_ROOT {

		size := this.matterDao.SizeByPuuidAndSpaceUuid(MATTER_ROOT, space.Uuid)

		this.spaceDao.UpdateTotalSize(space.Uuid, size)

		//update user total size info in cache.
		space.TotalSize = size
	} else {

		//compute self.
		size := this.matterDao.SizeByPuuidAndSpaceUuid(dirMatter.Uuid, user.Uuid)

		//when changed, we update
		if dirMatter.Size != size {
			this.spaceDao.UpdateTotalSize(space.Uuid, size)
		}
	}

}

// inner create directory.
func (this *MatterService) createDirectory(request *http.Request, dirMatter *Matter, name string, user *User, space *Space) *Matter {

	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil"))
	}

	if !dirMatter.Dir {
		panic(result.BadRequest("dirMatter must be directory"))
	}

	if dirMatter.Deleted {
		panic(result.BadRequest("Dir has been deleted. Cannot create dir under it."))
	}

	if dirMatter.SpaceUuid != space.Uuid {

		panic(result.BadRequest("file's space not the same"))
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
	matter := this.matterDao.FindBySpaceNameAndPuuidAndDirAndName(space.Name, dirMatter.Uuid, TRUE, name)
	if matter != nil {
		return matter
	}

	parts := strings.Split(dirMatter.Path, "/")

	if len(parts) > MATTER_NAME_MAX_DEPTH {
		panic(result.BadRequestI18n(request, i18n.MatterDepthExceedLimit, len(parts), MATTER_NAME_MAX_DEPTH))
	}

	absolutePath := GetSpaceMatterRootDir(space.Name) + dirMatter.Path + "/" + name

	relativePath := dirMatter.Path + "/" + name

	//crate directory on disk.
	dirPath := util.MakeDirAll(absolutePath)
	this.logger.Info("Create Directory: %s", dirPath)

	//create in db
	matter = &Matter{
		Puuid:     dirMatter.Uuid,
		UserUuid:  user.Uuid,
		SpaceUuid: space.Uuid,
		SpaceName: space.Name,
		Dir:       true,
		Name:      name,
		Path:      relativePath,
		Privacy:   false,
		VisitTime: time.Now(),
	}

	matter = this.matterDao.Create(matter)

	return matter
}

func (this *MatterService) AtomicCreateDirectory(request *http.Request, dirMatter *Matter, name string, user *User, space *Space) *Matter {

	if dirMatter.Deleted {
		panic(result.BadRequest("Dir has been deleted. Cannot create sub dir under it."))
	}

	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	matter := this.createDirectory(request, dirMatter, name, user, space)

	return matter
}

// copy or move may overwrite.
func (this *MatterService) handleOverwrite(request *http.Request, user *User, space *Space, destinationPath string, overwrite bool) {

	destMatter := this.matterDao.findByUserUuidAndPath(user.Uuid, destinationPath)
	if destMatter != nil {
		//if exist
		if overwrite {
			//delete.
			this.Delete(request, destMatter, user, space)
		} else {
			//throw precondition failed. (RFC4918:10.6)
			panic(result.CustomWebResult(result.PRECONDITION_FAILED, fmt.Sprintf("%s exists", destMatter.Path)))
		}
	}

}

// move srcMatter to destMatter. invoker must handled the overwrite and lock.
func (this *MatterService) move(request *http.Request, srcMatter *Matter, destDirMatter *Matter, user *User, space *Space) {

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
	this.ComputeRouteSize(srcPuuid, user, space)
	this.ComputeRouteSize(destDirUuid, user, space)

}

// move srcMatter to destMatter(must be dir)
func (this *MatterService) AtomicMove(request *http.Request, srcMatter *Matter, destDirMatter *Matter, overwrite bool, user *User, space *Space) {

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
	this.handleOverwrite(request, user, space, destinationPath, overwrite)

	//do the move operation.
	this.move(request, srcMatter, destDirMatter, user, space)
}

// move srcMatters to destMatter(must be dir)
func (this *MatterService) AtomicMoveBatch(request *http.Request, srcMatters []*Matter, destDirMatter *Matter, user *User, space *Space) {

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
		this.move(request, srcMatter, destDirMatter, user, space)
	}

}

// copy srcMatter to destMatter. invoker must handled the overwrite and lock.
func (this *MatterService) copy(request *http.Request, srcMatter *Matter, destDirMatter *Matter, name string) {

	this.logger.Info("copy srcPath = %s destPath = %s/%s", srcMatter.Path, destDirMatter.Path, name)

	if srcMatter.Dir {

		newMatter := &Matter{
			Puuid:     destDirMatter.Uuid,
			UserUuid:  srcMatter.UserUuid,
			SpaceName: srcMatter.SpaceName,
			Dir:       srcMatter.Dir,
			Name:      name,
			Md5:       "",
			Size:      srcMatter.Size,
			Privacy:   srcMatter.Privacy,
			Path:      destDirMatter.Path + "/" + name,
			Prop:      EMPTY_JSON_MAP,
			VisitTime: time.Now(),
		}

		newMatter = this.matterDao.Create(newMatter)

		//make the dir
		util.MakeDirAll(newMatter.AbsolutePath())

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
			Puuid:     destDirMatter.Uuid,
			UserUuid:  srcMatter.UserUuid,
			SpaceName: srcMatter.SpaceName,
			Dir:       srcMatter.Dir,
			Name:      name,
			Md5:       "",
			Size:      srcMatter.Size,
			Privacy:   srcMatter.Privacy,
			Path:      destDirMatter.Path + "/" + name,
			Prop:      EMPTY_JSON_MAP,
			VisitTime: time.Now(),
		}
		newMatter = this.matterDao.Create(newMatter)

	}
}

// copy srcMatter to destMatter.
func (this *MatterService) AtomicCopy(request *http.Request, srcMatter *Matter, destDirMatter *Matter, name string, overwrite bool, user *User, space *Space) {

	if srcMatter == nil {
		panic(result.BadRequest("srcMatter cannot be nil."))
	}

	this.userService.MatterLock(srcMatter.UserUuid)
	defer this.userService.MatterUnlock(srcMatter.UserUuid)

	if !destDirMatter.Dir {
		panic(result.BadRequestI18n(request, i18n.MatterDestinationMustDirectory))
	}

	destinationPath := destDirMatter.Path + "/" + name
	this.handleOverwrite(request, user, space, destinationPath, overwrite)

	this.copy(request, srcMatter, destDirMatter, name)
}

// rename matter to name
func (this *MatterService) AtomicRename(request *http.Request, matter *Matter, name string, overwrite bool, user *User, space *Space) {

	this.logger.Info("Try to rename srcPath = %s to name = %s", matter.Path, name)

	if user == nil {
		panic(result.BadRequest("user cannot be nil"))
	}

	if matter.Deleted {
		panic(result.BadRequest("matter has been deleted. Cannot rename."))
	}

	this.userService.MatterLock(user.Uuid)
	defer this.userService.MatterUnlock(user.Uuid)

	name = CheckMatterName(request, name)

	if name == matter.Name {
		panic(result.BadRequestI18n(request, i18n.MatterNameNoChange))
	}

	//check whether the name used by another matter.
	oldMatter := this.matterDao.FindBySpaceNameAndPuuidAndDirAndName(space.Name, matter.Puuid, "", name)
	if oldMatter != nil {
		if overwrite {
			//delete this one.
			this.Delete(request, oldMatter, user, space)
		} else {
			panic(result.CustomWebResult(result.PRECONDITION_FAILED, fmt.Sprintf("%s already exists", name)))
		}

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

// 将本地文件映射到蓝眼云盘中去。
func (this *MatterService) AtomicMirror(request *http.Request, srcPath string, destPath string, overwrite bool, user *User, space *Space) {

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

	destDirMatter := this.CreateDirectories(request, user, space, destPath)

	if destDirMatter.Deleted {
		panic(result.BadRequest("dest matter has been deleted. Cannot mirror."))
	}

	this.mirror(request, srcPath, destDirMatter, overwrite, user, space)
}

// 将本地文件/文件夹映射到蓝眼云盘中去。
func (this *MatterService) mirror(request *http.Request, srcPath string, destDirMatter *Matter, overwrite bool, user *User, space *Space) {

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
		srcDirMatter := this.matterDao.FindBySpaceNameAndPuuidAndDirAndName(space.Name, destDirMatter.Uuid, TRUE, fileStat.Name())

		if srcDirMatter == nil {
			srcDirMatter = this.createDirectory(request, destDirMatter, fileStat.Name(), user, space)
		}

		fileInfos, err := ioutil.ReadDir(srcPath)
		this.PanicError(err)

		//递归处理本文件夹下的文件或文件夹
		for _, fileInfo := range fileInfos {

			path := fmt.Sprintf("%s/%s", srcPath, fileInfo.Name())
			this.mirror(request, path, srcDirMatter, overwrite, user, space)
		}

	} else {

		//判断当前文件夹下，文件是否已经存在了。
		matter := this.matterDao.FindBySpaceNameAndPuuidAndDirAndName(space.Name, destDirMatter.Uuid, FALSE, fileStat.Name())
		if matter != nil {
			//如果是覆盖，那么删除之前的文件
			if overwrite {
				this.Delete(request, matter, user, space)
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

		this.Upload(request, file, nil, user, space, destDirMatter, fileStat.Name(), true)

	}

}

// 根据一个文件夹路径，依次创建，找到最后一个文件夹的matter，如果中途出错，返回err. 如果存在了那就直接返回即可。
func (this *MatterService) CreateDirectories(request *http.Request, user *User, space *Space, dirPath string) *Matter {

	dirPath = path.Clean(dirPath)

	if dirPath == "/" {
		return NewRootMatter(space)
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
			dirMatter = NewRootMatter(space)
			continue
		}

		dirMatter = this.createDirectory(request, dirMatter, name, user, space)
	}

	return dirMatter
}

// wrap a matter. put its parent.
func (this *MatterService) WrapParentDetail(request *http.Request, matter *Matter) *Matter {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil."))
	}

	//when self not root.
	if matter.Uuid != MATTER_ROOT {

		puuid := matter.Puuid
		tmpMatter := matter
		for puuid != MATTER_ROOT {
			pFile := this.matterDao.CheckByUuid(puuid)
			tmpMatter.Parent = pFile
			tmpMatter = pFile
			puuid = pFile.Puuid
		}
	}

	return matter
}

// wrap a matter ,put its children
func (this *MatterService) WrapChildrenDetail(request *http.Request, matter *Matter) {

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil."))
	}

	if matter.Dir {

		children := this.matterDao.FindByPuuidAndSpaceUuid(matter.Uuid, matter.SpaceUuid, nil)
		matter.Children = children

		for _, child := range matter.Children {
			this.WrapChildrenDetail(request, child)
		}
	}

}

// fetch a matter's detail with parent info.
func (this *MatterService) Detail(request *http.Request, uuid string) *Matter {
	matter := this.matterDao.CheckByUuid(uuid)
	return this.WrapParentDetail(request, matter)
}

// crawl a url to dirMatter
func (this *MatterService) AtomicCrawl(request *http.Request, url string, filename string, user *User, space *Space, dirMatter *Matter, privacy bool) *Matter {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	if dirMatter.Deleted {
		panic(result.BadRequest("Dir has been deleted. Cannot crawl under it."))
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
	//if resp is not ok.
	if resp.StatusCode != 200 {
		panic(result.BadRequest("error when crawl from url."))
	}

	return this.Upload(request, resp.Body, nil, user, space, dirMatter, filename, privacy)
}

// adjust a matter's path.
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

// delete someone's EyeblueTank files according to physics files.
func (this *MatterService) DeleteByPhysics(request *http.Request, user *User, space *Space) {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	//scan user's file. scan level by level.
	rootMatter := NewRootMatter(space)
	this.deleteFolderByPhysics(request, rootMatter, user, space)

}

func (this *MatterService) deleteFolderByPhysics(request *http.Request, dirMatter *Matter, user *User, space *Space) {

	//scan user's file. scan level by level.
	this.matterDao.PageHandle(dirMatter.Uuid, "", space.Uuid, "", "", "", nil, nil, func(matter *Matter) {

		if matter.Dir {
			//delete children first.
			this.deleteFolderByPhysics(request, matter, user, space)
		}

		if !util.PathExists(matter.AbsolutePath()) {
			this.logger.Info("physics file not exist. delete from tank. %s", matter.Name)
			this.AtomicDelete(nil, matter, user, space)
		}

	})
}

// scan someone's physics files to EyeblueTank
func (this *MatterService) ScanPhysics(request *http.Request, user *User, space *Space) {

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	rootDirPath := GetSpaceMatterRootDir(user.Username)
	this.logger.Info("scan %s's root dir %s", user.Username, rootDirPath)

	rootExists := util.PathExists(rootDirPath)
	if !rootExists {
		util.MakeDirAll(rootDirPath)
	}
	rootFileInfo, err := os.Lstat(rootDirPath)
	if err != nil {
		panic(result.BadRequest("cannot get root file info."))
	}

	rootMatter := NewRootMatter(space)
	this.scanPhysicsFolder(request, rootFileInfo, rootMatter, user, space)
}

func (this *MatterService) scanPhysicsFolder(request *http.Request, dirInfo os.FileInfo, dirMatter *Matter, user *User, space *Space) {
	if !dirInfo.IsDir() {
		return
	}

	//fetch all matters under this folder.
	_, matters := this.matterDao.PlainPage(0, 1000, dirMatter.Uuid, "", space.Uuid, "", "", "", nil, nil, nil)
	nameMatterMap := make(map[string]*Matter)
	for _, m := range matters {
		nameMatterMap[m.Name] = m
	}

	dirPath := dirMatter.AbsolutePath()
	names, err := util.ReadDirNames(dirPath)
	if err != nil {
		this.logger.Error("occur error when ReadDirNames %s %s", dirPath, err.Error())
		return
	}
	for _, name := range names {
		fileFullPath := filepath.Join(dirPath, name)
		fileInfo, err := os.Lstat(fileFullPath)
		if err != nil {
			this.logger.Error("occur error when Lstat %s %s", name, err.Error())
			continue
		}

		//find ther matter
		var matter *Matter
		_, ok := nameMatterMap[name]
		if ok {
			//exits. check the basic info.
			matter = nameMatterMap[name]
			//only check the fileSize.
			if !matter.Dir {
				if matter.Size != fileInfo.Size() {
					this.logger.Info("update matter: %s size:%d -> %d", name, matter.Size, fileInfo.Size())
					this.updateNonDirMatter(matter, fileInfo.Size(), user, space)
				}
			} else {

				//recursive scan this folder.
				this.scanPhysicsFolder(request, fileInfo, matter, user, space)

			}

		} else {

			if fileInfo.IsDir() {

				//create folder.
				matter = this.createDirectory(request, dirMatter, name, user, space)

				//recursive scan this folder.
				this.scanPhysicsFolder(request, fileInfo, matter, user, space)

			} else {

				//not exist. add basic info.
				this.logger.Info("Create matter: %s size:%d", name, fileInfo.Size())
				matter = this.createNonDirMatter(dirMatter, name, fileInfo.Size(), true, user, space)

			}

		}
	}
}

// clean all the expired deleted matters
func (this *MatterService) CleanExpiredDeletedMatters() {
	//mock a request.
	request := &http.Request{}

	preference := this.preferenceService.Fetch()

	this.userDao.PageHandle("", "", func(user *User, space *Space) {

		this.logger.Info("Clean %s 's deleted matters", user.Username)

		thenDate := time.Now()
		thenDate = thenDate.AddDate(0, 0, int(-preference.DeletedKeepDays))
		if preference.DeletedKeepDays != 0 {
			thenDate = util.FirstSecondOfDay(thenDate)
		}

		//first remove all the matter(not dir).
		this.matterDao.PageHandle("", "", "", "", FALSE, TRUE, &thenDate, nil, func(matter *Matter) {
			this.Delete(request, matter, user, space)
		})

		sortArray := []builder.OrderPair{
			{
				Key:   "path",
				Value: DIRECTION_DESC,
			},
		}

		//remove all the deleted directories. sort by path.
		this.matterDao.PageHandle("", "", "", "", TRUE, TRUE, &thenDate, sortArray, func(matter *Matter) {
			this.Delete(request, matter, user, space)
		})

	})

}
