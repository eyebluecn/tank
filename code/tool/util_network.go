package tool

import (
	"net/http"
	"strings"
)

//根据一个请求，获取ip.
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

//根据一个请求，获取host
func GetHostFromRequest(r *http.Request) string {

	return r.Host

}

//根据一个请求，获取authenticationId
func GetSessionUuidFromRequest(request *http.Request, cookieAuthKey string) string {

	//验证用户是否已经登录。
	sessionCookie, err := request.Cookie(cookieAuthKey)
	var sessionId string
	if err != nil {
		//从入参中捞取
		sessionId = request.FormValue(cookieAuthKey)
	} else {
		sessionId = sessionCookie.Value
	}

	return sessionId

}

//允许跨域请求
func AllowCORS(writer http.ResponseWriter) {
	writer.Header().Add("Access-Control-Allow-Origin", "*")
	writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
	writer.Header().Add("Access-Control-Max-Age", "3600")
	writer.Header().Add("Access-Control-Allow-Headers", "content-type")
}

//禁用缓存
func DisableCache(writer http.ResponseWriter) {
	//对于IE浏览器，会自动缓存，因此设置不缓存Header.
	writer.Header().Set("Pragma", "No-cache")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Expires", "0")
}
