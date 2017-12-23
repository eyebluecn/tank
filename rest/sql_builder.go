package rest

type OrderPair struct {
	key   string
	value string
}

type WherePair struct {
	Query string
	Args  []interface{}
}

func (this *WherePair) And(where *WherePair) *WherePair {
	if this.Query == "" {
		return where
	} else {
		return &WherePair{Query: this.Query + " AND " + where.Query, Args: append(this.Args, where.Args...)}
	}

}

func (this *WherePair) Or(where *WherePair) *WherePair {
	if this.Query == "" {
		return where
	} else {
		return &WherePair{Query: this.Query + " OR " + where.Query, Args: append(this.Args, where.Args...)}
	}

}

//根据一个sortMap，获取到order字符串
func (this *BaseDao) GetSortString(sortArray []OrderPair) string {

	if sortArray == nil || len(sortArray) == 0 {
		return ""
	}
	str := ""
	for _, pair := range sortArray {
		if pair.value == "DESC" || pair.value == "ASC" {
			if str != "" {
				str = str + ","
			}
			str = str + " " + pair.key + " " + pair.value
		}
	}

	return str
}
