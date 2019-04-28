package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

//@Service
type ShareService struct {
	BaseBean
	shareDao *ShareDao
	userDao  *UserDao
}

//初始化方法
func (this *ShareService) Init() {
	this.BaseBean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个分享的详情。
func (this *ShareService) Detail(uuid string) *Share {

	share := this.shareDao.CheckByUuid(uuid)

	return share
}
