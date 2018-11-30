package rest

/**
 * 系统的所有访问记录均记录在此
 */
type Footprint struct {
	Base
	UserUuid string `json:"userUuid"`
	Ip       string `json:"ip"`
	Host     string `json:"host"`
	Uri      string `json:"uri"`
	Params   string `json:"params"`
	Cost     int64  `json:"cost"`
	Success  bool   `json:"success"`
	Dt       string `json:"dt"`
}

// set File's table name to be `profiles`
func (Footprint) TableName() string {
	return TABLE_PREFIX + "footprint"
}
