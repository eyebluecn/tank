package rest

import "github.com/eyebluecn/tank/code/core"

//@Service
type SessionService struct {
	BaseBean
	userDao    *UserDao
	sessionDao *SessionDao
}

//初始化方法
func (this *SessionService) Init() {
	this.BaseBean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.sessionDao)
	if b, ok := b.(*SessionDao); ok {
		this.sessionDao = b
	}

}

//执行清理操作
func (this *SessionService) Cleanup() {

	this.logger.Info("[SessionService]执行清理：清除缓存中所有Session记录，共%d条。", core.CONTEXT.GetSessionCache().Count())

	core.CONTEXT.GetSessionCache().Truncate()
}
