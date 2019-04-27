package rest

import (
	"encoding/json"
	"github.com/eyebluecn/tank/code/config"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/robfig/cron"
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
		Ip:      util.GetIpAddress(request),
		Host:    request.Host,
		Uri:     request.URL.Path,
		Params:  paramsString,
		Cost:    int64(duration / time.Millisecond),
		Success: success,
	}

	//有可能DB尚且没有配置 直接打印出内容，并且退出
	if config.CONFIG.Installed {
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

//系统启动，数据库配置完毕后会调用该方法
func (this *FootprintService) Bootstrap() {

	//每日00:10 删除8日之前的访问数据
	expression := "0 10 0 * * ?"
	cronJob := cron.New()
	err := cronJob.AddFunc(expression, this.cleanOldData)
	util.PanicError(err)
	cronJob.Start()
	this.logger.Info("[cron job] 每日00:10 删除8日之前的访问数据")

	//立即执行一次数据清洗任务
	go util.SafeMethod(this.cleanOldData)

}

//定期删除8日前的数据。
func (this *FootprintService) cleanOldData() {

	day8Ago := time.Now()
	day8Ago = day8Ago.AddDate(0, 0, -8)
	day8Ago = util.FirstSecondOfDay(day8Ago)

	this.logger.Info("删除%s之前的访问数据", util.ConvertTimeToDateTimeString(day8Ago))

	this.footprintDao.DeleteByCreateTimeBefore(day8Ago)
}
