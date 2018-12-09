package rest

import (
	"encoding/json"
	"net/http"
	"time"
)

//@Service
type FootprintService struct {
	Bean
	footprintDao *FootprintDao
	userDao      *UserDao
}

//初始化方法
func (this *FootprintService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *FootprintService) Detail(uuid string) *Footprint {

	footprint := this.footprintDao.CheckByUuid(uuid)

	return footprint
}

//记录访问记录
func (this *FootprintService) Trace(writer http.ResponseWriter, request *http.Request, duration time.Duration, success bool) {

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
	footprint := &Footprint{
		Ip:      GetIpAddress(request),
		Host:    request.Host,
		Uri:     request.URL.Path,
		Params:  paramsString,
		Cost:    int64(duration / time.Millisecond),
		Success: success,
	}

	//有可能DB尚且没有配置 直接打印出内容，并且退出
	if CONFIG.Installed {
		user := this.findUser(writer, request)
		userUuid := ""
		if user != nil {
			userUuid = user.Uuid
		}
		footprint.UserUuid = userUuid
		footprint = this.footprintDao.Create(footprint)
	}

	//用json的方式输出返回值。
	this.logger.Info("Ip:%s Host:%s Uri:%s Params:%s Cost:%d", footprint.Ip, footprint.Host, footprint.Uri, paramsString, int64(duration/time.Millisecond))

}
