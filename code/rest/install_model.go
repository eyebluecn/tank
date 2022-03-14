package rest

// InstallTableInfo /**
// table meta info.
type InstallTableInfo struct {
	Name          string              `json:"name"`
	TableExist    bool                `json:"tableExist"`
	AllFields     []*InstallFieldInfo `json:"allFields"`
	MissingFields []*InstallFieldInfo `json:"missingFields"`
}

/**
 * table meta info.
 */
type InstallFieldInfo struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}
