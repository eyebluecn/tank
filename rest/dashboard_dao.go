package rest

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nu7hatch/gouuid"
	"time"
)

type DashboardDao struct {
	BaseDao
}

//创建
func (this *DashboardDao) Create(dashboard *Dashboard) *Dashboard {

	timeUUID, _ := uuid.NewV4()
	dashboard.Uuid = string(timeUUID.String())
	dashboard.CreateTime = time.Now()
	dashboard.UpdateTime = time.Now()
	db := CONTEXT.DB.Create(dashboard)
	this.PanicError(db.Error)

	return dashboard
}

//修改一条记录
func (this *DashboardDao) Save(dashboard *Dashboard) *Dashboard {

	dashboard.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(dashboard)
	this.PanicError(db.Error)

	return dashboard
}


//删除一条记录
func (this *DashboardDao) Delete(dashboard *Dashboard) {

	db := CONTEXT.DB.Delete(&dashboard)
	this.PanicError(db.Error)
}


//按照dt查询
func (this *DashboardDao) FindByDt(dt string) *Dashboard {

	// Read
	var dashboard Dashboard
	db := CONTEXT.DB.Where(&Dashboard{Dt: dt}).First(&dashboard)
	if db.Error != nil {
		return nil
	}
	return &dashboard
}
