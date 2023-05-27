package util

import (
	"net/http"
	"strconv"
)

// param is required. when missing, panic error.
func ExtractRequestString(request *http.Request, key string, errorHint string) string {
	str := request.FormValue(key)
	if str == "" {
		panic(errorHint)
	} else {
		return str
	}
}

// param is required. when missing, panic error.
func ExtractRequestInt64(request *http.Request, key string, errorHint string) int64 {
	keyStr := request.FormValue(key)

	var num int64 = 0
	if keyStr == "" {
		panic(errorHint)
	} else {
		intVal, err := strconv.Atoi(keyStr)
		if err != nil {
			panic(err)
		}
		num = int64(intVal)
		return num
	}
}

// param is required. when missing, panic error.
func ExtractRequestOptionalInt(request *http.Request, key string, defaultValue int) int {
	str := request.FormValue(key)
	if str == "" {
		return defaultValue
	} else {
		intVal, err := strconv.Atoi(str)
		if err != nil {
			panic(err)
		}
		return intVal
	}
}

// param is required. when missing, panic error.
func ExtractRequestOptionalString(request *http.Request, key string, defaultValue string) string {
	str := request.FormValue(key)
	if str == "" {
		return defaultValue
	} else {
		return str
	}
}
