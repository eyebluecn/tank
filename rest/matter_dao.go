package rest

import (
	"fmt"
	"github.com/jinzhu/gorm"

	"github.com/nu7hatch/gouuid"
	"os"
	"strings"
	"time"
)

type MatterDao struct {
	BaseDao
	imageCacheDao *ImageCacheDao
}

//初始化方法
func (this *MatterDao) Init() {
	this.BaseDao.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}
}

//按照Id查询文件
func (this *MatterDao) FindByUuid(uuid string) *Matter {

	// Read
	var matter Matter
	db := CONTEXT.DB.Where(&Matter{Base: Base{Uuid: uuid}}).First(&matter)
	if db.Error != nil {
		return nil
	}
	return &matter
}

//按照Id查询文件
func (this *MatterDao) CheckByUuid(uuid string) *Matter {

	// Read
	var matter Matter
	db := CONTEXT.DB.Where(&Matter{Base: Base{Uuid: uuid}}).First(&matter)
	this.PanicError(db.Error)

	return &matter

}

//按照名字查询文件夹
func (this *MatterDao) FindByUserUuidAndPuuidAndNameAndDirTrue(userUuid string, puuid string, name string) *Matter {

	var wp = &WherePair{}

	if userUuid != "" {
		wp = wp.And(&WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if puuid != "" {
		wp = wp.And(&WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if name != "" {
		wp = wp.And(&WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	wp = wp.And(&WherePair{Query: "dir = ?", Args: []interface{}{1}})

	var matter = &Matter{}
	db := CONTEXT.DB.Model(&Matter{}).Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		return nil
	}

	return matter
}

//按照id和userUuid来查找。找不到抛异常。
func (this *MatterDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *Matter {

	// Read
	var matter = &Matter{}
	db := CONTEXT.DB.Where(&Matter{Base: Base{Uuid: uuid}, UserUuid: userUuid}).First(matter)
	this.PanicError(db.Error)

	return matter

}

//统计某个用户的某个文件夹下的某个名字的文件(或文件夹)数量。
func (this *MatterDao) CountByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) int {

	var matter Matter
	var count int

	var wp = &WherePair{}

	if puuid != "" {
		wp = wp.And(&WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	wp = wp.And(&WherePair{Query: "dir = ?", Args: []interface{}{dir}})

	db := CONTEXT.DB.
		Model(&matter).
		Where(wp.Query, wp.Args...).
		Count(&count)
	this.PanicError(db.Error)

	return count
}

//获取某个用户的某个文件夹下的某个名字的文件(或文件夹)列表
func (this *MatterDao) ListByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) []*Matter {

	var matters []*Matter

	db := CONTEXT.DB.
		Where(Matter{UserUuid: userUuid, Puuid: puuid, Dir: dir, Name: name}).
		Find(&matters)
	this.PanicError(db.Error)

	return matters
}

//获取某个文件夹下所有的文件和子文件
func (this *MatterDao) List(puuid string, userUuid string, sortArray []OrderPair) []*Matter {
	var matters []*Matter

	db := CONTEXT.DB.Where(Matter{UserUuid: userUuid, Puuid: puuid}).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}

//获取某个文件夹下所有的文件和子文件
func (this *MatterDao) Page(page int, pageSize int, puuid string, userUuid string, name string, dir string, alien string, extensions []string, sortArray []OrderPair) *Pager {

	var wp = &WherePair{}

	if puuid != "" {
		wp = wp.And(&WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&WherePair{Query: "name LIKE ?", Args: []interface{}{"%" + name + "%"}})
	}

	if dir == "true" {
		wp = wp.And(&WherePair{Query: "dir = ?", Args: []interface{}{1}})
	} else if dir == "false" {
		wp = wp.And(&WherePair{Query: "dir = ?", Args: []interface{}{0}})
	}

	if alien == "true" {
		wp = wp.And(&WherePair{Query: "alien = ?", Args: []interface{}{1}})
	} else if alien == "false" {
		wp = wp.And(&WherePair{Query: "alien = ?", Args: []interface{}{0}})
	}

	var conditionDB *gorm.DB
	if extensions != nil && len(extensions) > 0 {
		var orWp = &WherePair{}

		for _, v := range extensions {
			orWp = orWp.Or(&WherePair{Query: "name LIKE ?", Args: []interface{}{"%." + v}})
		}

		conditionDB = CONTEXT.DB.Model(&Matter{}).Where(wp.Query, wp.Args...).Where(orWp.Query, orWp.Args...)
	} else {
		conditionDB = CONTEXT.DB.Model(&Matter{}).Where(wp.Query, wp.Args...)
	}

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var matters []*Matter
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&matters)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, matters)

	return pager
}

//创建
func (this *MatterDao) Create(matter *Matter) *Matter {

	timeUUID, _ := uuid.NewV4()
	matter.Uuid = string(timeUUID.String())
	matter.CreateTime = time.Now()
	matter.UpdateTime = time.Now()
	matter.Sort = time.Now().UnixNano() / 1e6
	db := CONTEXT.DB.Create(matter)
	this.PanicError(db.Error)

	return matter
}

//修改一个文件
func (this *MatterDao) Save(matter *Matter) *Matter {

	matter.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(matter)
	this.PanicError(db.Error)

	return matter
}

//计数器加一
func (this *MatterDao) TimesIncrement(matterUuid string) {
	db := CONTEXT.DB.Model(&Matter{}).Where("uuid = ?", matterUuid).Update("times", gorm.Expr("times + 1"))
	this.PanicError(db.Error)
}

//删除一个文件，数据库中删除，物理磁盘上删除。
func (this *MatterDao) Delete(matter *Matter) {

	//目录的话递归删除。
	if matter.Dir {
		matters := this.List(matter.Uuid, matter.UserUuid, nil)

		for _, f := range matters {
			this.Delete(f)
		}

		//删除文件夹本身
		db := CONTEXT.DB.Delete(&matter)
		this.PanicError(db.Error)

	} else {
		//删除数据库中文件记录
		db := CONTEXT.DB.Delete(&matter)
		this.PanicError(db.Error)

		//删除对应的缓存图片。
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		filePath := CONFIG.MatterPath + matter.Path
		//递归找寻文件的上级目录uuid. 因为是/开头的缘故
		parts := strings.Split(matter.Path, "/")
		dirPath := CONFIG.MatterPath + "/" + parts[1] + "/" + parts[2] + "/" + parts[3]

		//删除文件
		err := os.Remove(filePath)
		if err != nil {
			this.logger.Error(fmt.Sprintf("删除磁盘上的文件出错 %s", err.Error()))
		}

		//删除这一层文件夹
		err = os.Remove(dirPath)
		if err != nil {
			this.logger.Error(fmt.Sprintf("删除磁盘上的文件夹出错 %s", err.Error()))
		}

	}
}

//获取一段时间中，总的数量
func (this *MatterDao) CountBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var count int64
	db := CONTEXT.DB.Model(&Matter{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Count(&count)
	this.PanicError(db.Error)
	return count
}

//获取一段时间中文件总大小
func (this *MatterDao) SizeBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var size int64
	db := CONTEXT.DB.Model(&Matter{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	row.Scan(&size)
	return size
}
