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
func (this *SpaceMemberService) CreateMember(space *Space, user *User, spaceRole string) *SpaceMember {

	spaceMember := &SpaceMember{
		SpaceUuid: space.Uuid,
		UserUuid:  user.Uuid,
		Role:      spaceRole,
	}

	spaceMember = this.spaceMemberDao.Create(spaceMember)

	return spaceMember

}

// 当前用户对于此空间，是否有管理权限。
func (this *SpaceMemberService) canManage(user *User, spaceUuid string) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}
	if user.SpaceUuid == spaceUuid {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canManageBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有可读权限。
func (this *SpaceMemberService) canRead(user *User, spaceUuid string) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}
	if user.SpaceUuid == spaceUuid {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canReadBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有可写权限。
func (this *SpaceMemberService) canWrite(user *User, spaceUuid string) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}
	if user.SpaceUuid == spaceUuid {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canWriteBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有管理权限。
func (this *SpaceMemberService) canManageBySpaceMember(user *User, member *SpaceMember) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil && member.Role == SPACE_MEMBER_ROLE_ADMIN {
		return true
	}

	return false
}

// 当前用户对于此空间，是否有可读权限。
func (this *SpaceMemberService) canReadBySpaceMember(user *User, member *SpaceMember) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil {
		return true
	}

	return false
}

// 当前用户对于此空间，是否有科协权限。
func (this *SpaceMemberService) canWriteBySpaceMember(user *User, member *SpaceMember) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil && (member.Role == SPACE_MEMBER_ROLE_ADMIN || member.Role == SPACE_MEMBER_ROLE_READ_WRITE) {
		return true
	}

	return false
}
