package rest

import (
	"github.com/jinzhu/gorm"
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
	dashboard.Sort = time.Now().UnixNano() / 1e6
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

//获取某个文件夹下所有的文件和子文件
func (this *DashboardDao) Page(page int, pageSize int, dt string, sortArray []OrderPair) *Pager {

	var wp = &WherePair{}

	if dt != "" {
		wp = wp.And(&WherePair{Query: "dt = ?", Args: []interface{}{dt}})
	}

	var conditionDB *gorm.DB
	conditionDB = CONTEXT.DB.Model(&Dashboard{}).Where(wp.Query, wp.Args...)

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var dashboards []*Dashboard
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&dashboards)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, dashboards)

	return pager
}

//获取最活跃的前10个ip
func (this *DashboardDao) ActiveIpTop10() []*DashboardIpTimes {

	var dashboardIpTimes []*DashboardIpTimes

	sortArray := []OrderPair{
		{
			key:   "times",
			value: "DESC",
		},
	}
	rows, err := CONTEXT.DB.Model(&Footprint{}).
		Select("ip,COUNT(uuid) as times").
		Group("ip").
		Order(this.GetSortString(sortArray)).
		Offset(0).
		Limit(10).
		Rows()

	this.PanicError(err)
	for rows.Next() {
		var ip string;
		var times int64 = 0;
		rows.Scan(&ip, &times)
		item := &DashboardIpTimes{
			Ip:    ip,
			Times: times,
		}
		dashboardIpTimes = append(dashboardIpTimes, item)
	}

	return dashboardIpTimes
}
