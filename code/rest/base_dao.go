package rest

import (
	"github.com/eyebluecn/tank/code/tool/builder"
)

type BaseDao struct {
	BaseBean
}

//get an order string by sortMap
func (this *BaseDao) GetSortString(sortArray []builder.OrderPair) string {

	if sortArray == nil || len(sortArray) == 0 {
		return ""
	}
	str := ""
	for _, pair := range sortArray {
		if pair.Value == DIRECTION_DESC || pair.Value == DIRECTION_ASC {
			if str != "" {
				str = str + ","
			}
			str = str + " " + pair.Key + " " + pair.Value
		}
	}

	return str
}
