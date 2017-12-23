package rest

type Matter struct {
	Base
	Puuid    string  `json:"puuid"`
	UserUuid string  `json:"userUuid"`
	Dir      bool    `json:"dir"`
	Name     string  `json:"name"`
	Md5      string  `json:"md5"`
	Size     int64   `json:"size"`
	Privacy  bool    `json:"privacy"`
	Path     string  `json:"path"`
	Parent   *Matter `gorm:"-" json:"parent"`
}

// set File's table name to be `profiles`
func (Matter) TableName() string {
	return TABLE_PREFIX + "matter"
}
