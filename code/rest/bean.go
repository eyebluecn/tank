package rest

import (
	"github.com/eyebluecn/tank/code/config"
	"github.com/eyebluecn/tank/code/tool/inter"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
)

type IBean interface {
	//初始化方法
	Init()
	//系统清理方法
	Cleanup()
	//所有配置都加载完成后调用的方法，包括数据库加载完毕
	Bootstrap()
	//快速的Panic方法
	PanicError(err error)
}

type Bean struct {
	logger inter.Logger
}

func (this *Bean) Init() {
	this.logger = inter.LOGGER
}

func (this *Bean) Bootstrap() {

}

//系统大清理，一般时产品即将上线时，清除脏数据，只执行一次。
func (this *Bean) Cleanup() {

}

//处理错误的统一方法 可以省去if err!=nil 这段代码
func (this *Bean) PanicError(err error) {
	util.PanicError(err)
}

//能找到一个user就找到一个
func (this *Bean) findUser(writer http.ResponseWriter, request *http.Request) *User {

	//验证用户是否已经登录。
	//登录身份有效期以数据库中记录的为准
	sessionId := util.GetSessionUuidFromRequest(request, config.COOKIE_AUTH_KEY)
	if sessionId == "" {
		return nil
	}

	//去缓存中捞取看看
	cacheItem, err := CONTEXT.SessionCache.Value(sessionId)
	if err != nil {
		this.logger.Warn("获取缓存时出错了" + err.Error())
		return nil
	}

	if cacheItem == nil || cacheItem.Data() == nil {

		this.logger.Warn("cache item中已经不存在了 ")
		return nil
	}

	if value, ok := cacheItem.Data().(*User); ok {
		return value
	} else {
		this.logger.Error("cache item中的类型不是*User ")
	}

	return nil
}

//获取当前登录的用户，找不到就返回登录错误
func (this *Bean) checkUser(writer http.ResponseWriter, request *http.Request) *User {
	if this.findUser(writer, request) == nil {
		panic(result.ConstWebResult(result.CODE_WRAPPER_LOGIN))
	} else {
		return this.findUser(writer, request)
	}
}
