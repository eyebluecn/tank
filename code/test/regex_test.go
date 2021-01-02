package test

import (
	"regexp"
	"testing"
)

func TestUsername(t *testing.T) {

	testMap := make(map[string]bool)
	testMap[`tank`] = true
	testMap[`孙悟空`] = true
	testMap[`孙悟wukong`] = true
	testMap[`孙悟八戒`] = true
	testMap[`孙悟123`] = true
	testMap[`西天123`] = true
	testMap[`西天-123`] = false
	testMap[`西天@123`] = false
	testMap[`-西天@123`] = false
	testMap[`hong hua`] = false

	for k, v := range testMap {
		pattern := "^[\\p{Han}0-9a-zA-Z_]+$"
		usernameRegex := regexp.MustCompile(pattern)
		//使用MatchString来将要匹配的字符串传到匹配规则中
		result := usernameRegex.MatchString(k)

		if v == result {
			t.Logf(" %s = %v pass", k, v)
		} else {
			t.Errorf(" %s != %v error", k, v)
		}
	}

}
