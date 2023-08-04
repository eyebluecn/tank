package rest

import "time"

/**
 * application's dashboard.
 */
type Dashboard struct {
	Uuid           string    `json:"uuid" gorm:"type:char(36);primary_key;unique"`
	Sort           int64     `json:"sort" gorm:"type:bigint(20) not null"`
	UpdateTime     time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	CreateTime     time.Time `json:"createTime" gorm:"type:timestamp not null;default:'2018-01-01 00:00:00'"`
	InvokeNum      int64     `json:"invokeNum" gorm:"type:bigint(20) not null"`                  //api invoke num.
	TotalInvokeNum int64     `json:"totalInvokeNum" gorm:"type:bigint(20) not null;default:0"`   //total invoke num up to now.
	Uv             int64     `json:"uv" gorm:"type:bigint(20) not null;default:0"`               //today's uv
	TotalUv        int64     `json:"totalUv" gorm:"type:bigint(20) not null;default:0"`          //total uv
	MatterNum      int64     `json:"matterNum" gorm:"type:bigint(20) not null;default:0"`        //file's num
	TotalMatterNum int64     `json:"totalMatterNum" gorm:"type:bigint(20) not null;default:0"`   //file's total number
	FileSize       int64     `json:"fileSize" gorm:"type:bigint(20) not null;default:0"`         //today's file size
	TotalFileSize  int64     `json:"totalFileSize" gorm:"type:bigint(20) not null;default:0"`    //total file's size
	AvgCost        int64     `json:"avgCost" gorm:"type:bigint(20) not null;default:0"`          //api time cost in ms
	Dt             string    `json:"dt" gorm:"type:varchar(45) not null;index:idx_dashboard_dt"` //date. index should unique globally for sqlite.
}

/**
 * ip
 */
type DashboardIpTimes struct {
	Ip    string `json:"ip"`
	Times int64  `json:"times"`
}
