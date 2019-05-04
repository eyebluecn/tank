package rest

import "github.com/eyebluecn/tank/code/core"

//@Service
type SessionService struct {
	BaseBean
	userDao    *UserDao
	sessionDao *SessionDao
}

func (this *SessionService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

	b = core.CONTEXT.GetBean(this.sessionDao)
	if b, ok := b.(*SessionDao); ok {
		this.sessionDao = b
	}

}

//System cleanup.
func (this *SessionService) Cleanup() {

	this.logger.Info("[SessionService] clean up. Delete all Session. total:%d", core.CONTEXT.GetSessionCache().Count())

	core.CONTEXT.GetSessionCache().Truncate()
}
