package rest

import "github.com/eyebluecn/tank/code/tool/builder"

type BaseDao struct {
	BaseBean
}

//根据一个sortMap，获取到order字符串
func (this *BaseDao) GetSortString(sortArray []builder.OrderPair) string {

	if sortArray == nil || len(sortArray) == 0 {
		return ""
	}
	str := ""
	for _, pair := range sortArray {
		if pair.Value == "DESC" || pair.Value == "ASC" {
			if str != "" {
				str = str + ","
			}
			str = str + " " + pair.Key + " " + pair.Value
		}
	}

	return str
}
