package rest

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nu7hatch/gouuid"
	"time"
)

type SecurityVisitDao struct {
	BaseDao
}

//按照Id查询文件
func (this *SecurityVisitDao) FindByUuid(uuid string) *SecurityVisit {

	// Read
	var securityVisit SecurityVisit
	db := this.context.DB.Where(&SecurityVisit{Base: Base{Uuid: uuid}}).First(&securityVisit)
	if db.Error != nil {
		return nil
	}
	return &securityVisit
}

//按照Id查询文件
func (this *SecurityVisitDao) CheckByUuid(uuid string) *SecurityVisit {

	// Read
	var securityVisit SecurityVisit
	db := this.context.DB.Where(&SecurityVisit{Base: Base{Uuid: uuid}}).First(&securityVisit)
	this.PanicError(db.Error)

	return &securityVisit

}

//按分页条件获取分页
func (this *SecurityVisitDao) Page(page int, pageSize int, userUuid string, sortArray []OrderPair) *Pager {

	var wp = &WherePair{}

	if userUuid != "" {
		wp = wp.And(&WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	var conditionDB *gorm.DB
	conditionDB = this.context.DB.Model(&SecurityVisit{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var securityVisits []*SecurityVisit
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&securityVisits)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, securityVisits)

	return pager
}

//创建
func (this *SecurityVisitDao) Create(securityVisit *SecurityVisit) *SecurityVisit {

	timeUUID, _ := uuid.NewV4()
	securityVisit.Uuid = string(timeUUID.String())
	securityVisit.CreateTime = time.Now()
	securityVisit.UpdateTime = time.Now()
	db := this.context.DB.Create(securityVisit)
	this.PanicError(db.Error)

	return securityVisit
}

//修改一条记录
func (this *SecurityVisitDao) Save(securityVisit *SecurityVisit) *SecurityVisit {

	securityVisit.UpdateTime = time.Now()
	db := this.context.DB.Save(securityVisit)
	this.PanicError(db.Error)

	return securityVisit
}

//删除一条记录
func (this *SecurityVisitDao) Delete(securityVisit *SecurityVisit) {

	db := this.context.DB.Delete(&securityVisit)
	this.PanicError(db.Error)
}
