package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//@Service
type SecurityVisitService struct {
	Bean
	securityVisitDao *SecurityVisitDao
	userDao          *UserDao
}

//初始化方法
func (this *SecurityVisitService) Init() {

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.securityVisitDao)
	if b, ok := b.(*SecurityVisitDao); ok {
		this.securityVisitDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *SecurityVisitService) Detail(uuid string) *SecurityVisit {

	securityVisit := this.securityVisitDao.CheckByUuid(uuid)

	return securityVisit
}



//记录访问记录
func (this *SecurityVisitService) Log(writer http.ResponseWriter, request *http.Request) {
	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	var securityVisitDao *SecurityVisitDao
	b := CONTEXT.GetBean(securityVisitDao)
	if b, ok := b.(*SecurityVisitDao); ok {
		securityVisitDao = b
	}

	fmt.Printf("Host = %s Uri = %s  Path = %s  RawPath = %s  RawQuery = %s \n",
		request.Host,
		request.RequestURI,
		request.URL.Path,
		request.URL.RawPath,
		request.URL.RawQuery)

	params := make(map[string][]string)

	//POST请求参数
	values := request.PostForm
	for key, val := range values {
		params[key] = val
	}
	//GET请求参数
	values1 := request.URL.Query()
	for key, val := range values1 {
		params[key] = val
	}

	//用json的方式输出返回值。
	paramsString := "{}"
	paramsData, err := json.Marshal(params)
	if err == nil {
		paramsString = string(paramsData)
	}

	//将文件信息存入数据库中。
	securityVisit := &SecurityVisit{
		SessionId: "",
		UserUuid:  "testUserUUid",
		Ip:        GetIpAddress(request),
		Host:      request.Host,
		Uri:       request.URL.Path,
		Params:    paramsString,
		Cost:      0,
		Success:   true,
	}

	securityVisit = securityVisitDao.Create(securityVisit)

}
