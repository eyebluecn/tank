package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

//@Service
type BridgeService struct {
	BaseBean
	bridgeDao *BridgeDao
	userDao   *UserDao
}

//初始化方法
func (this *BridgeService) Init() {
	this.BaseBean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *BridgeService) Detail(uuid string) *Bridge {

	bridge := this.bridgeDao.CheckByUuid(uuid)

	return bridge
}
