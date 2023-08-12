package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
)

type SpaceController struct {
	BaseController
	spaceDao           *SpaceDao
	spaceMemberDao     *SpaceMemberDao
	spaceMemberService *SpaceMemberService
	matterDao          *MatterDao
	matterService      *MatterService
	spaceService       *SpaceService
	userService        *UserService
}

func (this *SpaceController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberDao)
	if b, ok := b.(*SpaceMemberDao); ok {
		this.spaceMemberDao = b
	}

	b = core.CONTEXT.GetBean(this.spaceMemberService)
	if b, ok := b.(*SpaceMemberService); ok {
		this.spaceMemberService = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.spaceService)
	if b, ok := b.(*SpaceService); ok {
		this.spaceService = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
	}

}

func (this *SpaceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/space/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/edit"] = this.Wrap(this.Edit, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/delete"] = this.Wrap(this.Delete, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/space/page"] = this.Wrap(this.Page, USER_ROLE_USER)
	return routeMap
}

func (this *SpaceController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	//space's name
	name := util.ExtractRequestString(request, "name")
	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	//create related space.
	space := this.spaceService.CreateSpace(request, name, nil, sizeLimit, totalSizeLimit, SPACE_TYPE_SHARED)

	return this.Success(space)
}

func (this *SpaceController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//space's uuid
	uuid := util.ExtractRequestString(request, "uuid")
	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	user := this.checkUser(request)
	space := this.spaceService.Edit(request, user, uuid, sizeLimit, totalSizeLimit)

	return this.Success(space)
}

func (this *SpaceController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//space's name
	uuid := util.ExtractRequestString(request, "uuid")
	space := this.spaceDao.CheckByUuid(uuid)

	//when space has members, cannot delete.
	memberCount := this.spaceMemberDao.CountBySpaceUuid(uuid)
	if memberCount > 0 {
		panic(result.BadRequest("space has members, cannot be deleted."))
	}

	//TODO: when space has files, cannot delete.

	//delete the space.
	this.spaceDao.Delete(space)

	return this.Success(nil)
}

func (this *SpaceController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := util.ExtractRequestString(request, "uuid")

	user := this.checkUser(request)
	space := this.spaceDao.CheckByUuid(uuid)
	canRead := this.spaceMemberService.canRead(user, space.Uuid)
	if !canRead {
		panic(result.BadRequestI18n(request, i18n.PermissionDenied))
	}

	return this.Success(space)

}

func (this *SpaceController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	page := util.ExtractRequestOptionalInt(request, "page", 0)
	pageSize := util.ExtractRequestOptionalInt(request, "pageSize", 20)
	orderCreateTime := util.ExtractRequestOptionalString(request, "orderCreateTime", "")
	spaceType := util.ExtractRequestOptionalString(request, "type", "")
	name := util.ExtractRequestOptionalString(request, "name", "")

	user := this.checkUser(request)

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
	}

	var pager *Pager
	if user.Role == USER_ROLE_USER {
		if spaceType != SPACE_TYPE_SHARED {
			panic(result.BadRequest("user can only query shared space type."))
		}
		pager = this.spaceDao.SelfPage(page, pageSize, user.Uuid, spaceType, sortArray)
	} else if user.Role == USER_ROLE_ADMINISTRATOR {
		pager = this.spaceDao.Page(page, pageSize, spaceType, name, sortArray)
	}

	return this.Success(pager)
}
