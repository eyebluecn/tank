package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/jinzhu/gorm"

	"github.com/nu7hatch/gouuid"
	"time"
)

type BridgeDao struct {
	BaseDao
}

//按照Id查询文件
func (this *BridgeDao) FindByUuid(uuid string) *Bridge {

	// Read
	var bridge Bridge
	db := core.CONTEXT.GetDB().Where(&Bridge{Base: Base{Uuid: uuid}}).First(&bridge)
	if db.Error != nil {
		return nil
	}
	return &bridge
}

//按照Id查询文件
func (this *BridgeDao) CheckByUuid(uuid string) *Bridge {

	// Read
	var bridge Bridge
	db := core.CONTEXT.GetDB().Where(&Bridge{Base: Base{Uuid: uuid}}).First(&bridge)
	this.PanicError(db.Error)

	return &bridge

}

//按照shareUuid和matterUuid查找
func (this *BridgeDao) CheckByShareUuidAndMatterUuid(shareUuid string, matterUuid string) *Bridge {

	// Read
	var bridge Bridge
	db := core.CONTEXT.GetDB().Where("share_uuid = ? AND matter_uuid = ?", shareUuid, matterUuid).First(&bridge)
	this.PanicError(db.Error)

	return &bridge

}

//按分页条件获取分页
func (this *BridgeDao) Page(page int, pageSize int, shareUuid string, sortArray []builder.OrderPair) *Pager {

	var wp = &builder.WherePair{}

	if shareUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "share_uuid = ?", Args: []interface{}{shareUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&Bridge{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var bridges []*Bridge
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&bridges)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, bridges)

	return pager
}

//创建
func (this *BridgeDao) Create(bridge *Bridge) *Bridge {

	timeUUID, _ := uuid.NewV4()
	bridge.Uuid = string(timeUUID.String())
	bridge.CreateTime = time.Now()
	bridge.UpdateTime = time.Now()
	bridge.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(bridge)
	this.PanicError(db.Error)

	return bridge
}

//修改一条记录
func (this *BridgeDao) Save(bridge *Bridge) *Bridge {

	bridge.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(bridge)
	this.PanicError(db.Error)

	return bridge
}

//删除一条记录
func (this *BridgeDao) Delete(bridge *Bridge) {

	db := core.CONTEXT.GetDB().Delete(&bridge)
	this.PanicError(db.Error)
}

//删除一个matter对应的所有缓存
func (this *BridgeDao) DeleteByMatterUuid(matterUuid string) {

	var wp = &builder.WherePair{}

	wp = wp.And(&builder.WherePair{Query: "matter_uuid = ?", Args: []interface{}{matterUuid}})

	//删除文件记录
	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args).Delete(Bridge{})
	this.PanicError(db.Error)
}

//删除一个share对应的所有缓存
func (this *BridgeDao) DeleteByShareUuid(shareUuid string) {

	var wp = &builder.WherePair{}

	wp = wp.And(&builder.WherePair{Query: "share_uuid = ?", Args: []interface{}{shareUuid}})

	//删除文件记录
	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args).Delete(Bridge{})
	this.PanicError(db.Error)
}

//根据shareUuid获取关联的所有matter.
func (this *BridgeDao) ListByShareUuid(shareUuid string) []*Bridge {

	if shareUuid == "" {
		panic(result.BadRequest("shareUuid cannot be nil"))
	}

	var bridges []*Bridge

	db := core.CONTEXT.GetDB().
		Where("share_uuid = ?", shareUuid).
		Find(&bridges)
	this.PanicError(db.Error)

	return bridges
}

//执行清理操作
func (this *BridgeDao) Cleanup() {
	this.logger.Info("[BridgeDao]执行清理：清除数据库中所有Bridge记录。")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Bridge{})
	this.PanicError(db.Error)
}
