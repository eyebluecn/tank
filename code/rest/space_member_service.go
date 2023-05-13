package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

// @Service
type SpaceMemberService struct {
	BaseBean
	spaceMemberDao *SpaceMemberDao
	matterDao      *MatterDao
	bridgeDao      *BridgeDao
	userDao        *UserDao
}

func (this *SpaceMemberService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*SpaceMemberDao); ok {
		this.spaceMemberDao = b
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

func (this *SpaceMemberService) Detail(uuid string) *SpaceMember {

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	return spaceMember
}

// create space
func (this *SpaceMemberService) CreateMember(space *Space, memeber *User, spaceRole string) *SpaceMember {

	spaceMember := &SpaceMember{
		SpaceUuid: space.Uuid,
		UserUuid:  memeber.Uuid,
		Role:      spaceRole,
	}

	spaceMember = this.spaceMemberDao.Create(spaceMember)

	return spaceMember

}
