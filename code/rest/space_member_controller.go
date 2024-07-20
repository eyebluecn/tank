package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"strings"
)

type SpaceMemberController struct {
	BaseController
	spaceMemberDao     *SpaceMemberDao
	spaceDao           *SpaceDao
	bridgeDao          *BridgeDao
	matterDao          *MatterDao
	matterService      *MatterService
	spaceMemberService *SpaceMemberService
}

func (this *SpaceMemberController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*SpaceMemberService); ok {
		this.spaceMemberService = b
	}

}

func (this *SpaceMemberController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//admin user can create/edit/delete
	routeMap["/api/space/member/create"] = this.Wrap(this.Create, USER_ROLE_USER)
	routeMap["/api/space/member/edit"] = this.Wrap(this.Edit, USER_ROLE_USER)
	routeMap["/api/space/member/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)

	routeMap["/api/space/member/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/space/member/mine"] = this.Wrap(this.Mine, USER_ROLE_USER)
	routeMap["/api/space/member/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

func (this *SpaceMemberController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	spaceUuid := util.ExtractRequestString(request, "spaceUuid")
	userUuidsStr := util.ExtractRequestString(request, "userUuids")
	spaceRole := util.ExtractRequestString(request, "role")

	if spaceRole != SPACE_MEMBER_ROLE_READ_ONLY && spaceRole != SPACE_MEMBER_ROLE_READ_WRITE && spaceRole != SPACE_MEMBER_ROLE_ADMIN {
		panic("spaceRole is not correct")
	}

	//validate userUuids
	if userUuidsStr == "" {
		panic("userUuids is required")
	}
	userUuids := strings.Split(userUuidsStr, ",")

	// check operator's permission
	currentUser := this.checkUser(request)
	canManage := this.spaceMemberService.canManage(currentUser, spaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	//check whether exists.
	for _, userUuid := range userUuids {
		spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, userUuid)
		user := this.userDao.CheckByUuid(userUuid)
		if spaceMember != nil {
			panic(result.BadRequestI18n(request, i18n.SpaceMemberExist, user.Username))
		}
	}

	//check whether space exists.
	space := this.spaceDao.CheckByUuid(spaceUuid)

	//check whether exists.
	for _, userUuid := range userUuids {
		user := this.userDao.CheckByUuid(userUuid)
		this.spaceMemberService.CreateMember(space, user, spaceRole)
	}

	return this.Success("OK")
}

func (this *SpaceMemberController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuid := util.ExtractRequestString(request, "uuid")
	spaceRole := util.ExtractRequestString(request, "role")

	if spaceRole != SPACE_MEMBER_ROLE_READ_ONLY && spaceRole != SPACE_MEMBER_ROLE_READ_WRITE && spaceRole != SPACE_MEMBER_ROLE_ADMIN {
		panic("spaceRole is not correct")
	}

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	currentUser := this.checkUser(request)
	canManage := this.spaceMemberService.canManage(currentUser, spaceMember.SpaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	spaceMember.Role = spaceRole
	spaceMember = this.spaceMemberDao.Save(spaceMember)

	return this.Success(spaceMember)
}

func (this *SpaceMemberController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	uuid := util.ExtractRequestString(request, "uuid")

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)
	user := this.checkUser(request)
	canManage := this.spaceMemberService.canManage(user, spaceMember.SpaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	this.spaceMemberDao.Delete(spaceMember)

	return this.Success("OK")
}

func (this *SpaceMemberController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	user := this.checkUser(request)

	if spaceMember.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(spaceMember)

}

// find my role in the space.
func (this *SpaceMemberController) Mine(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	spaceUuid := util.ExtractRequestString(request, "spaceUuid")

	user := this.checkUser(request)
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	if spaceMember == nil {
		spaceMember = &SpaceMember{SpaceUuid: spaceUuid, Role: SPACE_MEMBER_GUEST}
	}
	if user.Role == USER_ROLE_ADMINISTRATOR {
		spaceMember.Role = SPACE_MEMBER_ROLE_ADMIN
	}

	return this.Success(spaceMember)

}

func (this *SpaceMemberController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	page := util.ExtractRequestOptionalInt(request, "page", 0)
	pageSize := util.ExtractRequestOptionalInt(request, "pageSize", 20)
	orderCreateTime := util.ExtractRequestOptionalString(request, "orderCreateTime", "")
	spaceUuid := util.ExtractRequestString(request, "spaceUuid")

	user := this.checkUser(request)
	canRead := this.spaceMemberService.canRead(user, spaceUuid)
	if !canRead {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
	}

	pager := this.spaceMemberDao.Page(page, pageSize, spaceUuid, sortArray)

	//fill the space's user. FIXME: user better way to get User.
	if pager != nil {
		for _, spaceMember := range pager.Data.([]*SpaceMember) {
			spaceMember.User = this.userDao.FindByUuid(spaceMember.UserUuid)
		}
	}

	return this.Success(pager)
}
