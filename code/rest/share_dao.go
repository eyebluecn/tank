package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/jinzhu/gorm"

	"github.com/nu7hatch/gouuid"
	"time"
)

type ShareDao struct {
	BaseDao
}

//按照Id查询文件
func (this *ShareDao) FindByUuid(uuid string) *Share {

	// Read
	var share Share
	db := core.CONTEXT.GetDB().Where(&Share{Base: Base{Uuid: uuid}}).First(&share)
	if db.Error != nil {
		return nil
	}
	return &share
}

//按照Id查询文件
func (this *ShareDao) CheckByUuid(uuid string) *Share {

	// Read
	var share Share
	db := core.CONTEXT.GetDB().Where(&Share{Base: Base{Uuid: uuid}}).First(&share)
	this.PanicError(db.Error)

	return &share

}

//按分页条件获取分页
func (this *ShareDao) Page(page int, pageSize int, userUuid string, sortArray []builder.OrderPair) *Pager {

	var wp = &builder.WherePair{}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&Share{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var shares []*Share
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&shares)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, shares)

	return pager
}

//创建
func (this *ShareDao) Create(share *Share) *Share {

	timeUUID, _ := uuid.NewV4()
	share.Uuid = string(timeUUID.String())
	share.CreateTime = time.Now()
	share.UpdateTime = time.Now()
	share.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(share)
	this.PanicError(db.Error)

	return share
}

//修改一条记录
func (this *ShareDao) Save(share *Share) *Share {

	share.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(share)
	this.PanicError(db.Error)

	return share
}

//删除一条记录
func (this *ShareDao) Delete(share *Share) {

	db := core.CONTEXT.GetDB().Delete(&share)
	this.PanicError(db.Error)

}

//执行清理操作
func (this *ShareDao) Cleanup() {
	this.logger.Info("[ShareDao]执行清理：清除数据库中所有Share记录。")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Share{})
	this.PanicError(db.Error)
}
