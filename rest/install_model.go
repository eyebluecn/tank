package rest

/**
 * 表名对应的表结构
 */
type InstallTableInfo struct {
	Name           string `json:"name"`
	CreateSql      string `json:"createSql"`
	TableExist     bool   `json:"tableExist"`
	ExistCreateSql string `json:"existCreateSql"`
}
