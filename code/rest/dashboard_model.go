package rest

import "github.com/eyebluecn/tank/code/core"

/**
 * application's dashboard.
 */
type Dashboard struct {
	Base
	InvokeNum      int64  `json:"invokeNum" gorm:"type:bigint(20) not null"`                //api invoke num.
	TotalInvokeNum int64  `json:"totalInvokeNum" gorm:"type:bigint(20) not null;default:0"` //total invoke num up to now.
	Uv             int64  `json:"uv" gorm:"type:bigint(20) not null;default:0"`             //today's uv
	TotalUv        int64  `json:"totalUv" gorm:"type:bigint(20) not null;default:0"`        //total uv
	MatterNum      int64  `json:"matterNum" gorm:"type:bigint(20) not null;default:0"`      //file's num
	TotalMatterNum int64  `json:"totalMatterNum" gorm:"type:bigint(20) not null;default:0"` //file's total number
	FileSize       int64  `json:"fileSize" gorm:"type:bigint(20) not null;default:0"`       //today's file size
	TotalFileSize  int64  `json:"totalFileSize" gorm:"type:bigint(20) not null;default:0"`  //total file's size
	AvgCost        int64  `json:"avgCost" gorm:"type:bigint(20) not null;default:0"`        //api time cost in ms
	Dt             string `json:"dt" gorm:"type:varchar(45) not null;index:idx_dt"`         //date
}

// set File's table name to be `profiles`
func (this *Dashboard) TableName() string {
	return core.TABLE_PREFIX + "dashboard"
}

/**
 * ip
 */
type DashboardIpTimes struct {
	Ip    string `json:"ip"`
	Times int64  `json:"times"`
}
