package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"regexp"
)

// @Service
type SpaceService struct {
	BaseBean
	spaceDao           *SpaceDao
	spaceMemberService *SpaceMemberService
	matterDao          *MatterDao
	bridgeDao          *BridgeDao
	userDao            *UserDao
}

func (this *SpaceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*SpaceMemberService); ok {
		this.spaceMemberService = b
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
func (this *SpaceService) CreateSpace(
	request *http.Request,
	name string,
	user *User,
	sizeLimit int64,
	totalSizeLimit int64,
	spaceType string) *Space {

	userUuid := ""
	//validation work.
	if m, _ := regexp.MatchString(USERNAME_PATTERN, name); !m {
		panic(result.BadRequestI18n(request, i18n.SpaceNameError))
	}

	if spaceType == SPACE_TYPE_PRIVATE {
		if user == nil {
			panic("private space requires user.")
		}

		userUuid = user.Uuid
		if this.spaceDao.CountByUserUuid(userUuid) > 0 {
			panic(result.BadRequestI18n(request, i18n.SpaceExclusive, name))
		}

	} else if spaceType == SPACE_TYPE_SHARED {

	} else {
		panic("Not supported spaceType:" + spaceType)
	}

	if this.spaceDao.CountByName(name) > 0 {
		panic(result.BadRequestI18n(request, i18n.SpaceNameExist, name))
	}

	space := &Space{
		Name:           name,
		UserUuid:       userUuid,
		SizeLimit:      sizeLimit,
		TotalSizeLimit: totalSizeLimit,
		TotalSize:      0,
		Type:           spaceType,
	}

	space = this.spaceDao.Create(space)

	return space

}

// checkout a adminAble space.
func (this *SpaceService) CheckAdminAbleByUuid(request *http.Request, user *User, spaceUuid string) *Space {
	space := this.spaceDao.CheckByUuid(spaceUuid)
	if space.Type == SPACE_TYPE_PRIVATE && user.Uuid == space.UserUuid {
		return space
	}

	manage := this.spaceMemberService.canManage(user, spaceUuid)
	if !manage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return space
}

// checkout a writable space.
func (this *SpaceService) CheckWritableByUuid(request *http.Request, user *User, spaceUuid string) *Space {
	space := this.spaceDao.CheckByUuid(spaceUuid)
	if space.Type == SPACE_TYPE_PRIVATE && user.Uuid == space.UserUuid {
		return space
	}

	writable := this.spaceMemberService.canWrite(user, spaceUuid)
	if !writable {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return space
}

// checkout a readable space.
func (this *SpaceService) CheckReadableByUuid(request *http.Request, user *User, spaceUuid string) *Space {
	space := this.spaceDao.CheckByUuid(spaceUuid)
	if space.Type == SPACE_TYPE_PRIVATE && user.Uuid == space.UserUuid {
		return space
	}

	manage := this.spaceMemberService.canRead(user, spaceUuid)
	if !manage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return space
}

// edit space's info
func (this *SpaceService) Edit(request *http.Request, user *User, spaceUuid string, sizeLimit int64, totalSizeLimit int64) *Space {
	space := this.CheckAdminAbleByUuid(request, user, spaceUuid)

	if sizeLimit < 0 && sizeLimit != -1 {
		panic("sizeLimit cannot be negative expect -1.")
	}

	if totalSizeLimit < 0 && totalSizeLimit != -1 {
		panic("totalSizeLimit cannot be negative expect -1.")
	}

	space.SizeLimit = sizeLimit
	space.TotalSizeLimit = totalSizeLimit
	space = this.spaceDao.Save(space)

	return space
}
