package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

// @Service
type SpaceService struct {
	BaseBean
	spaceDao  *SpaceDao
	matterDao *MatterDao
	bridgeDao *BridgeDao
	userDao   *UserDao
}

func (this *SpaceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

func (this *SpaceService) Detail(uuid string) *Space {

	space := this.spaceDao.CheckByUuid(uuid)

	return space
}

// create space
func (this *SpaceService) CreateSpace(userUuid string) *Space {

	space := &Space{
		UserUuid: userUuid,
	}

	space = this.spaceDao.Create(space)

	return space

}
