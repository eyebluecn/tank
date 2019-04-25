package code

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
	"os"
	"path/filepath"
	"tank/code/tool/util"
	"time"
)

type ImageCacheDao struct {
	BaseDao
}

//按照Id查询文件
func (this *ImageCacheDao) FindByUuid(uuid string) *ImageCache {

	// Read
	var imageCache ImageCache
	db := CONTEXT.DB.Where(&ImageCache{Base: Base{Uuid: uuid}}).First(&imageCache)
	if db.Error != nil {
		return nil
	}
	return &imageCache
}

//按照Id查询文件
func (this *ImageCacheDao) CheckByUuid(uuid string) *ImageCache {

	// Read
	var imageCache ImageCache
	db := CONTEXT.DB.Where(&ImageCache{Base: Base{Uuid: uuid}}).First(&imageCache)
	this.PanicError(db.Error)

	return &imageCache

}

//按照名字查询文件夹
func (this *ImageCacheDao) FindByMatterUuidAndMode(matterUuid string, mode string) *ImageCache {

	var wp = &WherePair{}

	if matterUuid != "" {
		wp = wp.And(&WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})
	}

	if mode != "" {
		wp = wp.And(&WherePair{Query: "mode = ?", Args: []interface{}{mode}})
	}

	var imageCache = &ImageCache{}
	db := CONTEXT.DB.Model(&ImageCache{}).Where(wp.Query, wp.Args...).First(imageCache)

	if db.Error != nil {
		return nil
	}

	return imageCache
}

//按照id和userUuid来查找。找不到抛异常。
func (this *ImageCacheDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *ImageCache {

	// Read
	var imageCache = &ImageCache{}
	db := CONTEXT.DB.Where(&ImageCache{Base: Base{Uuid: uuid}, UserUuid: userUuid}).First(imageCache)
	this.PanicError(db.Error)

	return imageCache

}

//获取某个用户的某个文件夹下的某个名字的文件(或文件夹)列表
func (this *ImageCacheDao) ListByUserUuidAndPuuidAndDirAndName(userUuid string) []*ImageCache {

	var imageCaches []*ImageCache

	db := CONTEXT.DB.
		Where(ImageCache{UserUuid: userUuid}).
		Find(&imageCaches)
	this.PanicError(db.Error)

	return imageCaches
}

//获取某个文件夹下所有的文件和子文件
func (this *ImageCacheDao) Page(page int, pageSize int, userUuid string, matterUuid string, sortArray []OrderPair) *Pager {

	var wp = &WherePair{}

	if userUuid != "" {
		wp = wp.And(&WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if matterUuid != "" {
		wp = wp.And(&WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = CONTEXT.DB.Model(&ImageCache{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var imageCaches []*ImageCache
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&imageCaches)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, imageCaches)

	return pager
}

//创建
func (this *ImageCacheDao) Create(imageCache *ImageCache) *ImageCache {

	timeUUID, _ := uuid.NewV4()
	imageCache.Uuid = string(timeUUID.String())
	imageCache.CreateTime = time.Now()
	imageCache.UpdateTime = time.Now()
	imageCache.Sort = time.Now().UnixNano() / 1e6
	db := CONTEXT.DB.Create(imageCache)
	this.PanicError(db.Error)

	return imageCache
}

//修改一个文件
func (this *ImageCacheDao) Save(imageCache *ImageCache) *ImageCache {

	imageCache.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(imageCache)
	this.PanicError(db.Error)

	return imageCache
}

//删除一个文件包括文件夹
func (this *ImageCacheDao) deleteFileAndDir(imageCache *ImageCache) {

	filePath := GetUserCacheRootDir(imageCache.Username) + imageCache.Path

	dirPath := filepath.Dir(filePath)

	//删除文件
	err := os.Remove(filePath)
	if err != nil {
		this.logger.Error(fmt.Sprintf("删除磁盘上的文件%s出错 %s", filePath, err.Error()))
	}

	//如果这一层文件夹是空的，那么删除文件夹本身。
	util.DeleteEmptyDirRecursive(dirPath)

}

//删除一个文件，数据库中删除，物理磁盘上删除。
func (this *ImageCacheDao) Delete(imageCache *ImageCache) {

	db := CONTEXT.DB.Delete(&imageCache)
	this.PanicError(db.Error)

	this.deleteFileAndDir(imageCache)

}

//删除一个matter对应的所有缓存
func (this *ImageCacheDao) DeleteByMatterUuid(matterUuid string) {

	var wp = &WherePair{}

	wp = wp.And(&WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})

	//查询出即将删除的图片缓存
	var imageCaches []*ImageCache
	db := CONTEXT.DB.Where(wp.Query, wp.Args).Find(&imageCaches)
	this.PanicError(db.Error)

	//删除文件记录
	db = CONTEXT.DB.Where(wp.Query, wp.Args).Delete(ImageCache{})
	this.PanicError(db.Error)

	//删除文件实体
	for _, imageCache := range imageCaches {
		this.deleteFileAndDir(imageCache)
	}

}

//获取一段时间中文件总大小
func (this *ImageCacheDao) SizeBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var size int64
	db := CONTEXT.DB.Model(&ImageCache{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	row.Scan(&size)
	return size
}


//执行清理操作
func (this *ImageCacheDao) Cleanup() {
	this.logger.Info("[ImageCacheDao]执行清理：清除数据库中所有ImageCache记录。")
	db := CONTEXT.DB.Where("uuid is not null").Delete(ImageCache{})
	this.PanicError(db.Error)
}

