package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
	"os"
	"path/filepath"
	"time"
)

type ImageCacheDao struct {
	BaseDao
}

//find by uuid. if not found return nil.
func (this *ImageCacheDao) FindByUuid(uuid string) *ImageCache {
	var entity = &ImageCache{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", uuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

//find by uuid. if not found panic NotFound error
func (this *ImageCacheDao) CheckByUuid(uuid string) *ImageCache {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity

}

func (this *ImageCacheDao) FindByMatterUuidAndMode(matterUuid string, mode string) *ImageCache {

	var wp = &builder.WherePair{}

	if matterUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})
	}

	if mode != "" {
		wp = wp.And(&builder.WherePair{Query: "mode = ?", Args: []interface{}{mode}})
	}

	var imageCache = &ImageCache{}
	db := core.CONTEXT.GetDB().Model(&ImageCache{}).Where(wp.Query, wp.Args...).First(imageCache)

	if db.Error != nil {
		return nil
	}

	return imageCache
}

func (this *ImageCacheDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *ImageCache {

	// Read
	var imageCache = &ImageCache{}
	db := core.CONTEXT.GetDB().Where(&ImageCache{Base: Base{Uuid: uuid}, UserUuid: userUuid}).First(imageCache)
	this.PanicError(db.Error)

	return imageCache

}

func (this *ImageCacheDao) FindByUserUuidAndPuuidAndDirAndName(userUuid string) []*ImageCache {

	var imageCaches []*ImageCache

	db := core.CONTEXT.GetDB().
		Where(ImageCache{UserUuid: userUuid}).
		Find(&imageCaches)
	this.PanicError(db.Error)

	return imageCaches
}

func (this *ImageCacheDao) Page(page int, pageSize int, userUuid string, matterUuid string, sortArray []builder.OrderPair) *Pager {

	var wp = &builder.WherePair{}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if matterUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&ImageCache{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var imageCaches []*ImageCache
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&imageCaches)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, imageCaches)

	return pager
}

func (this *ImageCacheDao) Create(imageCache *ImageCache) *ImageCache {

	timeUUID, _ := uuid.NewV4()
	imageCache.Uuid = string(timeUUID.String())
	imageCache.CreateTime = time.Now()
	imageCache.UpdateTime = time.Now()
	imageCache.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(imageCache)
	this.PanicError(db.Error)

	return imageCache
}

func (this *ImageCacheDao) Save(imageCache *ImageCache) *ImageCache {

	imageCache.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(imageCache)
	this.PanicError(db.Error)

	return imageCache
}

func (this *ImageCacheDao) deleteFileAndDir(imageCache *ImageCache) {

	filePath := GetUserCacheRootDir(imageCache.Username) + imageCache.Path

	dirPath := filepath.Dir(filePath)

	//delete file from disk.
	err := os.Remove(filePath)
	if err != nil {
		this.logger.Error(fmt.Sprintf("error while deleting %s from disk %s", filePath, err.Error()))
	}

	//if this level is empty. Delete the directory
	util.DeleteEmptyDirRecursive(dirPath)

}

//delete a file from db and disk.
func (this *ImageCacheDao) Delete(imageCache *ImageCache) {

	db := core.CONTEXT.GetDB().Delete(&imageCache)
	this.PanicError(db.Error)

	this.deleteFileAndDir(imageCache)

}

//delete all the cache of a matter.
func (this *ImageCacheDao) DeleteByMatterUuid(matterUuid string) {

	var wp = &builder.WherePair{}

	wp = wp.And(&builder.WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})

	var imageCaches []*ImageCache
	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args).Find(&imageCaches)
	this.PanicError(db.Error)

	//delete from db.
	db = core.CONTEXT.GetDB().Where(wp.Query, wp.Args).Delete(ImageCache{})
	this.PanicError(db.Error)

	//delete from disk.
	for _, imageCache := range imageCaches {
		this.deleteFileAndDir(imageCache)
	}

}

func (this *ImageCacheDao) SizeBetweenTime(startTime time.Time, endTime time.Time) int64 {

	var wp = &builder.WherePair{Query: "create_time >= ? AND create_time <= ?", Args: []interface{}{startTime, endTime}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&ImageCache{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var size int64
	db = core.CONTEXT.GetDB().Model(&ImageCache{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&size)
	this.PanicError(err)
	return size
}

//System cleanup.
func (this *ImageCacheDao) Cleanup() {
	this.logger.Info("[ImageCacheDao]clean up. Delete all ImageCache ")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(ImageCache{})
	this.PanicError(db.Error)
}
