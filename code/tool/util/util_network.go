package util

import (
	"net/http"
	"strings"
)

//get ip from request
func GetIpAddress(r *http.Request) string {
	var ipAddress string

	ipAddress = r.RemoteAddr

	if ipAddress != "" {
		ipAddress = strings.Split(ipAddress, ":")[0]
	}

	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(r.Header.Get(h), ",") {
			if ip != "" {
				ipAddress = ip
			}
		}
	}
	return ipAddress
}

//get host from request
func GetHostFromRequest(request *http.Request) string {

	return request.Host

}

//get cookieAuthKey from request.
func GetSessionUuidFromRequest(request *http.Request, cookieAuthKey string) string {

	//get from cookie
	sessionCookie, err := request.Cookie(cookieAuthKey)
	var sessionId string
	if err != nil {
		//try to get from Form
		sessionId = request.FormValue(cookieAuthKey)
	} else {
		sessionId = sessionCookie.Value
	}

	return sessionId

}

//allow cors.
func AllowCORS(writer http.ResponseWriter) {
	writer.Header().Add("Access-Control-Allow-Origin", "*")
	writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	writer.Header().Add("Access-Control-Max-Age", "3600")
	writer.Header().Add("Access-Control-Allow-Headers", "content-type")
}

//disable cache.
func DisableCache(writer http.ResponseWriter) {
	//IE browser will cache automatically. disable the cache.
	writer.Header().Set("Pragma", "No-cache")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Expires", "0")
}
