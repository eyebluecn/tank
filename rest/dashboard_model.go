package rest

/**
 * 系统的所有访问记录均记录在此
 */
type Dashboard struct {
	Base
	VisitNum       int64  `json:"visitNum"`
	TotalVisitNum  int64  `json:"totalVisitNum"`
	Uv             int64  `json:"uv"`
	TotalUv        int64  `json:"totalUv"`
	MatterNum      int64  `json:"matterNum"`
	TotalMatterNum int64  `json:"totalMatterNum"`
	FileSize       int64  `json:"fileSize"`
	TotalFileSize  int64  `json:"totalFileSize"`
	AvgCost        int64  `json:"avgCost"`
	Dt             string `json:"dt"`
}

// set File's table name to be `profiles`
func (Dashboard) TableName() string {
	return TABLE_PREFIX + "dashboard"
}

/**
 * 总调用量
 */
type DashboardInvoke struct {
	InvokeNum int64  `json:"invokeNum"`
	Uv        int64  `json:"uv"`
	Dt        string `json:"dt"`
}
