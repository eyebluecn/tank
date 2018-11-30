package rest

/**
 * 总调用量
 */
type DashboardInvoke struct {
	InvokeNum int64 `json:"invokeNum"`
	Dt        string `json:"dt"`
}
