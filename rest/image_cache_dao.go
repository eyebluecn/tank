package rest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nu7hatch/gouuid"
	"os"
	"time"
)

type ImageCacheDao struct {
	BaseDao
}

//按照Id查询文件
func (this *ImageCacheDao) FindByUuid(uuid string) *ImageCache {

	// Read
	var imageCache ImageCache
	db := this.context.DB.Where(&ImageCache{Base: Base{Uuid: uuid}}).First(&imageCache)
	if db.Error != nil {
		return nil
	}
	return &imageCache
}

//按照Id查询文件
func (this *ImageCacheDao) CheckByUuid(uuid string) *ImageCache {

	// Read
	var imageCache ImageCache
	db := this.context.DB.Where(&ImageCache{Base: Base{Uuid: uuid}}).First(&imageCache)
	this.PanicError(db.Error)

	return &imageCache

}

//按照名字查询文件夹
func (this *ImageCacheDao) FindByUri(uri string) *ImageCache {

	var wp = &WherePair{}

	wp = wp.And(&WherePair{Query: "uri = ?", Args: []interface{}{uri}})

	var imageCache = &ImageCache{}
	db := this.context.DB.Model(&ImageCache{}).Where(wp.Query, wp.Args...).First(imageCache)

	if db.Error != nil {
		return nil
	}

	return imageCache
}

//按照id和userUuid来查找。找不到抛异常。
func (this *ImageCacheDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *ImageCache {

	// Read
	var imageCache = &ImageCache{}
	db := this.context.DB.Where(&ImageCache{Base: Base{Uuid: uuid}, UserUuid: userUuid}).First(imageCache)
	this.PanicError(db.Error)

	return imageCache

}

//获取某个用户的某个文件夹下的某个名字的文件(或文件夹)列表
func (this *ImageCacheDao) ListByUserUuidAndPuuidAndDirAndName(userUuid string) []*ImageCache {

	var imageCaches []*ImageCache

	db := this.context.DB.
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
	conditionDB = this.context.DB.Model(&ImageCache{}).Where(wp.Query, wp.Args...)

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
	db := this.context.DB.Create(imageCache)
	this.PanicError(db.Error)

	return imageCache
}

//修改一个文件
func (this *ImageCacheDao) Save(imageCache *ImageCache) *ImageCache {

	imageCache.UpdateTime = time.Now()
	db := this.context.DB.Save(imageCache)
	this.PanicError(db.Error)

	return imageCache
}

//删除一个文件，数据库中删除，物理磁盘上删除。
func (this *ImageCacheDao) Delete(imageCache *ImageCache) {

	db := this.context.DB.Delete(&imageCache)
	this.PanicError(db.Error)

	//删除文件
	err := os.Remove(CONFIG.MatterPath + imageCache.Path)

	if err != nil {
		LogError(fmt.Sprintf("删除磁盘上的文件出错，不做任何处理"))
	}
}

//删除一个matter对应的所有缓存
func (this *ImageCacheDao) DeleteByMatterUuid(matterUuid string) {

	var wp = &WherePair{}

	wp = wp.And(&WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})

	//查询出即将删除的图片缓存
	var imageCaches []*ImageCache
	db := this.context.DB.Where(wp.Query, wp.Args).Find(&imageCaches)
	this.PanicError(db.Error)

	//删除文件记录
	db = this.context.DB.Where(wp.Query, wp.Args).Delete(ImageCache{})
	this.PanicError(db.Error)

	//删除文件实体
	for _, imageCache := range imageCaches {
		err := os.Remove(CONFIG.MatterPath + imageCache.Path)
		if err != nil {
			LogError(fmt.Sprintf("删除磁盘上的文件出错，不做任何处理"))
		}
	}

}
