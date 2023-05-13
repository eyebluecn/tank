package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
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

	routeMap["/api/space/member/create"] = this.Wrap(this.Create, USER_ROLE_USER)
	routeMap["/api/space/member/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/space/member/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/space/member/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

func (this *SpaceMemberController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	spaceUuid := request.FormValue("spaceUuid")
	userUuid := request.FormValue("userUuid")
	spaceRole := request.FormValue("spaceRole")

	if spaceUuid == "" {
		panic("spaceUuid is required")
	}

	if spaceRole != SPACE_MEMBER_ROLE_READ_ONLY && spaceRole != SPACE_MEMBER_ROLE_READ_WRITE && spaceRole != SPACE_MEMBER_ROLE_ADMIN {
		panic("spaceRole is not correct")
	}

	currentUser := this.checkUser(request)
	canManage := this.canManage(currentUser, spaceUuid)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	//check whether exists.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, userUuid)
	if spaceMember != nil {
		panic(result.BadRequestI18n(request, i18n.SpaceMemberExist))
	}

	//check whether space exists.
	space := this.spaceDao.CheckByUuid(spaceUuid)
	member := this.userDao.CheckByUuid(userUuid)
	//can not add a SPACE_USER as member.
	if member.Role == USER_ROLE_SPACE {
		panic(result.BadRequestI18n(request, i18n.SpaceMemberRoleConflict))
	}

	spaceMember = this.spaceMemberService.CreateMember(space, member, spaceRole)

	return this.Success(spaceMember)
}

func (this *SpaceMemberController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	spaceMemberUuid := request.FormValue("spaceMemberUuid")
	spaceMember := this.spaceMemberDao.CheckByUuid(spaceMemberUuid)
	user := this.checkUser(request)
	canManage := this.canManageBySpaceMember(user, spaceMember)
	if !canManage {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	this.spaceMemberDao.Delete(spaceMember)

	return this.Success("OK")
}

func (this *SpaceMemberController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	spaceMember := this.spaceMemberDao.CheckByUuid(uuid)

	user := this.checkUser(request)

	if spaceMember.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(spaceMember)

}

func (this *SpaceMemberController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")

	user := this.checkUser(request)

	var page int
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	pageSize := 200
	if pageSizeStr != "" {
		tmp, err := strconv.Atoi(pageSizeStr)
		if err == nil {
			pageSize = tmp
		}
	}

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
	}

	pager := this.spaceMemberDao.Page(page, pageSize, user.Uuid, sortArray)

	return this.Success(pager)
}

// 当前用户对于此空间，是否有管理权限。
func (this *SpaceMemberController) canManage(user *User, spaceUuid string) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	spaceMember := this.spaceMemberDao.FindBySpaceUuidAndUserUuid(spaceUuid, user.Uuid)
	return this.canManageBySpaceMember(user, spaceMember)
}

// 当前用户对于此空间，是否有管理权限。
func (this *SpaceMemberController) canManageBySpaceMember(user *User, member *SpaceMember) bool {
	if user.Role == USER_ROLE_ADMINISTRATOR {
		return true
	}

	//only space's admin can add member.
	if member != nil && member.Role == SPACE_MEMBER_ROLE_ADMIN {
		return true
	}

	return false
}
