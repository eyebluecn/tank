package rest

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type DashboardDao struct {
	BaseDao
}

//过去七天调用量
func (this *DashboardDao) InvokeList() []*DashboardInvoke {

	var tableName = Footprint{}.TableName()
	now := time.Now()
	startDate := now.AddDate(0, 0, -6)
	rows, err := CONTEXT.DB.Raw("SELECT count(uuid) as invoke_num,dt FROM "+tableName+" where dt>= ? and dt <= ? group by dt",
		ConvertTimeToDateString(startDate),
		ConvertTimeToDateString(now)).Rows()
	this.PanicError(err)
	defer rows.Close()

	var invokeMap = make(map[string]int64)
	var dashboardInvokes []*DashboardInvoke
	for rows.Next() {
		var invokeNum int64 = 0;
		var dt string;
		rows.Scan(&invokeNum, &dt)
		invokeMap[dt] = invokeNum
	}
	for i := -6; i <= 0; i++ {
		date := now.AddDate(0, 0, i)
		dt := ConvertTimeToDateString(date)
		var invokeNum int64 = 0
		v, ok := invokeMap[dt]
		if ok {
			invokeNum = v
		}

		dashboardInvokes = append(dashboardInvokes, &DashboardInvoke{
			InvokeNum: invokeNum,
			Dt:        dt,
		})
	}

	return dashboardInvokes
}
