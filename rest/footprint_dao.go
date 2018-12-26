package rest

import (
	"github.com/jinzhu/gorm"

	"github.com/nu7hatch/gouuid"
	"time"
)

type FootprintDao struct {
	BaseDao
}

//按照Id查询文件
func (this *FootprintDao) FindByUuid(uuid string) *Footprint {

	// Read
	var footprint Footprint
	db := CONTEXT.DB.Where(&Footprint{Base: Base{Uuid: uuid}}).First(&footprint)
	if db.Error != nil {
		return nil
	}
	return &footprint
}

//按照Id查询文件
func (this *FootprintDao) CheckByUuid(uuid string) *Footprint {

	// Read
	var footprint Footprint
	db := CONTEXT.DB.Where(&Footprint{Base: Base{Uuid: uuid}}).First(&footprint)
	this.PanicError(db.Error)

	return &footprint

}

//按分页条件获取分页
func (this *FootprintDao) Page(page int, pageSize int, userUuid string, sortArray []OrderPair) *Pager {

	var wp = &WherePair{}

	if userUuid != "" {
		wp = wp.And(&WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = CONTEXT.DB.Model(&Footprint{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var footprints []*Footprint
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&footprints)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, footprints)

	return pager
}

//创建
func (this *FootprintDao) Create(footprint *Footprint) *Footprint {

	timeUUID, _ := uuid.NewV4()
	footprint.Uuid = string(timeUUID.String())
	footprint.CreateTime = time.Now()
	footprint.UpdateTime = time.Now()
	footprint.Sort = time.Now().UnixNano() / 1e6
	db := CONTEXT.DB.Create(footprint)
	this.PanicError(db.Error)

	return footprint
}

//修改一条记录
func (this *FootprintDao) Save(footprint *Footprint) *Footprint {

	footprint.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(footprint)
	this.PanicError(db.Error)

	return footprint
}

//删除一条记录
func (this *FootprintDao) Delete(footprint *Footprint) {

	db := CONTEXT.DB.Delete(&footprint)
	this.PanicError(db.Error)
}

//获取一段时间中，总的数量
func (this *FootprintDao) CountBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var count int64
	db := CONTEXT.DB.Model(&Footprint{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Count(&count)
	this.PanicError(db.Error)
	return count
}

//获取一段时间中UV的数量
func (this *FootprintDao) UvBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var count int64
	db := CONTEXT.DB.Model(&Footprint{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Select("COUNT(DISTINCT(ip))")
	this.PanicError(db.Error)
	row := db.Row()
	row.Scan(&count)
	return count
}

//获取一段时间中平均耗时
func (this *FootprintDao) AvgCostBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var cost float64
	db := CONTEXT.DB.Model(&Footprint{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Select("AVG(cost)")
	this.PanicError(db.Error)
	row := db.Row()
	row.Scan(&cost)
	return int64(cost)
}

