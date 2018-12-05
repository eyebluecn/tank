package rest

/**
 * 系统的所有访问记录均记录在此
 */
type Dashboard struct {
	Base
	InvokeNum      int64  `json:"invokeNum" gorm:"type:bigint(20) not null"`                //当日访问量
	TotalInvokeNum int64  `json:"totalInvokeNum" gorm:"type:bigint(20) not null;default:0"` //截至目前总访问量
	Uv             int64  `json:"uv" gorm:"type:bigint(20) not null;default:0"`             //当日UV
	TotalUv        int64  `json:"totalUv" gorm:"type:bigint(20) not null;default:0"`        //截至目前总UV
	MatterNum      int64  `json:"matterNum" gorm:"type:bigint(20) not null;default:0"`      //文件数量
	TotalMatterNum int64  `json:"totalMatterNum" gorm:"type:bigint(20) not null;default:0"` //截至目前文件数量
	FileSize       int64  `json:"fileSize" gorm:"type:bigint(20) not null;default:0"`       //当日文件大小
	TotalFileSize  int64  `json:"totalFileSize" gorm:"type:bigint(20) not null;default:0"`  //截至目前文件总大小
	AvgCost        int64  `json:"avgCost" gorm:"type:bigint(20) not null;default:0"`        //请求平均耗时 ms
	Dt             string `json:"dt" gorm:"type:varchar(45) not null;index:idx_dt"`         //日期
}

// set File's table name to be `profiles`
func (Dashboard) TableName() string {
	return TABLE_PREFIX + "dashboard"
}

/**
 * 统计IP活跃数的
 */
type DashboardIpTimes struct {
	Ip    string `json:"ip"`
	Times int64  `json:"times"`
}
