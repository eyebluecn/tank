package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
	"time"
)

type DashboardDao struct {
	BaseDao
}

func (this *DashboardDao) Create(dashboard *Dashboard) *Dashboard {

	timeUUID, _ := uuid.NewV4()
	dashboard.Uuid = string(timeUUID.String())
	dashboard.CreateTime = time.Now()
	dashboard.UpdateTime = time.Now()
	dashboard.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(dashboard)
	this.PanicError(db.Error)

	return dashboard
}

func (this *DashboardDao) Save(dashboard *Dashboard) *Dashboard {

	dashboard.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(dashboard)
	this.PanicError(db.Error)

	return dashboard
}

func (this *DashboardDao) Delete(dashboard *Dashboard) {

	db := core.CONTEXT.GetDB().Delete(&dashboard)
	this.PanicError(db.Error)
}

func (this *DashboardDao) FindByDt(dt string) *Dashboard {

	var dashboard = &Dashboard{}
	db := core.CONTEXT.GetDB().Where("dt = ?", dt).First(dashboard)
	if db.Error != nil {
		return nil
	}
	return dashboard
}

func (this *DashboardDao) Page(page int, pageSize int, dt string, sortArray []builder.OrderPair) *Pager {

	var wp = &builder.WherePair{}

	if dt != "" {
		wp = wp.And(&builder.WherePair{Query: "dt = ?", Args: []interface{}{dt}})
	}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&Dashboard{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var dashboards []*Dashboard
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&dashboards)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, dashboards)

	return pager
}

func (this *DashboardDao) ActiveIpTop10() []*DashboardIpTimes {

	var dashboardIpTimes []*DashboardIpTimes

	sortArray := []builder.OrderPair{
		{
			Key:   "times",
			Value: "DESC",
		},
	}
	rows, err := core.CONTEXT.GetDB().Model(&Footprint{}).
		Select("ip,COUNT(uuid) as times").
		Group("ip").
		Order(this.GetSortString(sortArray)).
		Offset(0).
		Limit(10).
		Rows()

	this.PanicError(err)
	for rows.Next() {
		var ip string
		var times int64 = 0
		err := rows.Scan(&ip, &times)
		this.PanicError(err)
		item := &DashboardIpTimes{
			Ip:    ip,
			Times: times,
		}
		dashboardIpTimes = append(dashboardIpTimes, item)
	}

	return dashboardIpTimes
}

func (this *DashboardDao) Cleanup() {
	this.logger.Info("[DashboardDao] cleanup. Delete all Dashboard records")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Dashboard{})
	this.PanicError(db.Error)
}
