package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"os"
	"regexp"
	"strings"
)

//@Service
type PreferenceService struct {
	BaseBean
	preferenceDao *PreferenceDao
	preference    *Preference
	matterDao     *MatterDao
	matterService *MatterService
	userDao       *UserDao
	migrating     bool
}

func (this *PreferenceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*PreferenceDao); ok {
		this.preferenceDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
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

}

func (this *PreferenceService) Fetch() *Preference {

	if this.preference == nil {
		this.preference = this.preferenceDao.Fetch()
	}

	return this.preference
}

//清空单例配置。
func (this *PreferenceService) Reset() {

	this.preference = nil

}

//System cleanup.
func (this *PreferenceService) Cleanup() {

	this.logger.Info("[PreferenceService] clean up. Delete all preference")

	this.Reset()
}

//migrate 2.0's db data and file data to 3.0
func (this *PreferenceService) Migrate20to30(writer http.ResponseWriter, request *http.Request) {

	matterPath := request.FormValue("matterPath")

	if matterPath == "" {
		panic(result.BadRequest("matterPath required"))
	}

	this.logger.Info("start migrating from 2.0 to 3.0")

	//lock
	if this.migrating {
		panic(result.BadRequest("migrating work is processing"))
	} else {
		this.migrating = true
	}
	defer func() {
		this.migrating = false
	}()

	//delete all users with _20
	this.userDao.DeleteUsers20()

	migrateDashboardSql := "INSERT INTO `tank`.`tank30_download_token` ( `uuid`, `sort`, `update_time`, `create_time`, `user_uuid`, `matter_uuid`, `expire_time`, `ip` ) ( SELECT `uuid`, `sort`, `update_time`, `create_time`, `user_uuid`, `matter_uuid`, `expire_time`, `ip` FROM `tank`.`tank20_download_token`)"
	this.logger.Info(migrateDashboardSql)
	db := core.CONTEXT.GetDB().Exec(migrateDashboardSql)
	if db.Error != nil {
		this.logger.Error("%v", db.Error)
	}

	migrateDownloadTokenSql := "INSERT INTO `tank`.`tank30_dashboard` ( `uuid`, `sort`, `update_time`, `create_time`, `invoke_num`, `total_invoke_num`, `uv`, `total_uv`, `matter_num`, `total_matter_num`, `file_size`, `total_file_size`, `avg_cost`, `dt` ) ( SELECT `uuid`, `sort`, `update_time`, `create_time`, `invoke_num`, `total_invoke_num`, `uv`, `total_uv`, `matter_num`, `total_matter_num`, `file_size`, `total_file_size`, `avg_cost`, `dt` FROM `tank`.`tank20_dashboard` )"
	this.logger.Info(migrateDownloadTokenSql)
	db = core.CONTEXT.GetDB().Exec(migrateDownloadTokenSql)
	if db.Error != nil {
		this.logger.Error("%v", db.Error)
	}

	migrateMatterSql := "INSERT INTO `tank`.`tank30_matter` ( `uuid`, `sort`, `update_time`, `create_time`, `puuid`, `user_uuid`, `username`, `dir`, `name`, `md5`, `size`, `privacy`, `path`, `times` ) ( SELECT `uuid`, `sort`, `update_time`, `create_time`, `puuid`, `user_uuid`, '', `dir`, `name`, `md5`, `size`, `privacy`, `path`, `times` FROM `tank`.`tank20_matter` ) "
	this.logger.Info(migrateMatterSql)
	db = core.CONTEXT.GetDB().Exec(migrateMatterSql)
	if db.Error != nil {
		this.logger.Error("%v", db.Error)
	}

	migrateUploadTokenSql := "INSERT INTO `tank`.`tank30_upload_token` ( `uuid`, `sort`, `update_time`, `create_time`, `user_uuid`, `folder_uuid`, `matter_uuid`, `expire_time`, `filename`, `privacy`, `size`, `ip` ) ( SELECT `uuid`, `sort`, `update_time`, `create_time`, `user_uuid`, `folder_uuid`, `matter_uuid`, `expire_time`, `filename`, `privacy`, `size`, `ip` FROM `tank`.`tank20_upload_token` ) "
	this.logger.Info(migrateUploadTokenSql)
	db = core.CONTEXT.GetDB().Exec(migrateUploadTokenSql)
	if db.Error != nil {
		this.logger.Error("%v", db.Error)
	}

	//username in tank2.0 add _20.
	migrateUserSql := "INSERT INTO `tank`.`tank30_user` ( `uuid`, `sort`, `update_time`, `create_time`, `role`, `username`, `password`, `avatar_url`, `last_ip`, `last_time`, `size_limit`, `total_size_limit`, `total_size`, `status` ) ( SELECT `uuid`, `sort`, `update_time`, `create_time`, `role`, CONCAT(`username`,'_20') as `username`, `password`, `avatar_url`, `last_ip`, `last_time`, `size_limit`, -1, 0, `status` FROM `tank`.`tank20_user` )"
	this.logger.Info(migrateUserSql)
	db = core.CONTEXT.GetDB().Exec(migrateUserSql)
	if db.Error != nil {
		this.logger.Error("%v", db.Error)
	}

	//find all 2.0 users.
	users := this.userDao.FindUsers20()
	for _, user := range users {
		this.logger.Info("start handling matters for user %s %s", user.Uuid, user.Username)
		rootMatter := NewRootMatter(user)
		firstLevelMatters := this.matterDao.FindByPuuidAndUserUuid(MATTER_ROOT, user.Uuid, nil)
		for _, firstLevelMatter := range firstLevelMatters {
			this.HandleMatter20(request, matterPath, rootMatter, firstLevelMatter, user)
		}

		//adjust all the size.
		this.matterService.ComputeAllDirSize(user)
	}
}

//handle matter from 2.0
func (this *PreferenceService) HandleMatter20(request *http.Request, matterPath string, dirMatter *Matter, matter *Matter, user *User) {
	defer func() {
		if err := recover(); err != nil {
			this.logger.Warn("HandleMatter20 occur error %v when handle matter %s %s. Ignore the error and continue. \r\n", err, matter.Uuid, matter.Name)
		}
	}()

	this.logger.Info("start handling matter %s", matter.Name)

	if matter == nil {
		panic(result.BadRequest("matter cannot be nil."))
	}

	if user == nil {
		panic(result.BadRequest("user cannot be nil."))
	}

	if dirMatter == nil {
		panic(result.BadRequest("dirMatter cannot be nil"))
	}

	if !dirMatter.Dir {
		panic(result.BadRequest("dirMatter must be directory"))
	}

	if dirMatter.UserUuid != user.Uuid {

		panic(result.BadRequest("file's user not the same"))
	}

	name := matter.Name
	filename := name
	if name == "" {
		panic(result.BadRequest("name cannot be blank"))
	}

	if len(name) > MATTER_NAME_MAX_LENGTH {

		panic(result.BadRequestI18n(request, i18n.MatterNameLengthExceedLimit, len(name), MATTER_NAME_MAX_LENGTH))

	}

	//if directory. Create it.
	if matter.Dir {

		if m, _ := regexp.MatchString(MATTER_NAME_PATTERN, name); m {
			panic(result.BadRequestI18n(request, i18n.MatterNameContainSpecialChars))
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

		//change matter info.
		matter.Username = user.Username
		matter.Path = relativePath

		matter = this.matterDao.Save(matter)

		//handle its children.
		children := this.matterDao.FindByPuuidAndUserUuid(matter.Uuid, user.Uuid, nil)
		for _, child := range children {
			this.HandleMatter20(request, matterPath, matter, child, user)
		}

	} else {

		//if file. copy and adjust it.

		dirAbsolutePath := dirMatter.AbsolutePath()
		dirRelativePath := dirMatter.Path

		fileAbsolutePath := dirAbsolutePath + "/" + filename
		fileRelativePath := dirRelativePath + "/" + filename

		util.MakeDirAll(dirAbsolutePath)

		//if exist, panic it.
		exist := util.PathExists(fileAbsolutePath)
		if exist {
			this.logger.Error("%s exits, overwrite it.", fileAbsolutePath)
			removeError := os.Remove(fileAbsolutePath)
			this.PanicError(removeError)
		}

		srcAbsolutePath := fmt.Sprintf("%s%s", matterPath, matter.Path)

		//find the 2.0 disk file.
		fileSize := util.CopyFile(srcAbsolutePath, fileAbsolutePath)

		this.logger.Info("copy %s %v ", filename, util.HumanFileSize(fileSize))

		//update info.
		matter.Path = fileRelativePath
		matter.Username = user.Username

		matter = this.matterDao.Save(matter)

	}

}
