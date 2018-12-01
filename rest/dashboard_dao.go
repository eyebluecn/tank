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

	//过去几天
	var dayNum = 15;
	var tableName = Footprint{}.TableName()
	now := time.Now()
	startDate := now.AddDate(0, 0, 1-dayNum)
	rows, err := CONTEXT.DB.Raw("SELECT COUNT(uuid) AS invoke_num,COUNT(DISTINCT(ip)) AS uv,dt FROM "+tableName+" WHERE dt>= ? AND dt <= ? GROUP BY dt",
		ConvertTimeToDateString(startDate),
		ConvertTimeToDateString(now)).Rows()
	this.PanicError(err)
	defer rows.Close()

	var invokeMap = make(map[string]*DashboardInvoke)
	var dashboardInvokes []*DashboardInvoke
	for rows.Next() {
		var invokeNum int64 = 0;
		var uv int64 = 0;
		var dt string;
		rows.Scan(&invokeNum, &uv, &dt)
		invokeMap[dt] = &DashboardInvoke{
			InvokeNum: invokeNum,
			Uv:        uv,
			Dt:        dt,
		}
	}
	for i := 1 - dayNum; i <= 0; i++ {
		date := now.AddDate(0, 0, i)
		dt := ConvertTimeToDateString(date)
		v, ok := invokeMap[dt]
		if ok {
			dashboardInvokes = append(dashboardInvokes, v)
		} else {
			dashboardInvokes = append(dashboardInvokes, &DashboardInvoke{
				InvokeNum: 0,
				Uv:        0,
				Dt:        dt,
			})
		}
	}

	return dashboardInvokes
}
