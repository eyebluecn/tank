package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
	"strings"
)

type SpaceMemberController struct {
	BaseController
	spaceMemberDao     *SpaceMemberDao
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

	routeMap["/api/spaceMember/delete/batch"] = this.Wrap(this.DeleteBatch, USER_ROLE_USER)
	routeMap["/api/spaceMember/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/spaceMember/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

func (this *SpaceMemberController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("uuids cannot be null"))
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		imageCache := this.spaceMemberDao.FindByUuid(uuid)

		user := this.checkUser(request)
		if imageCache.UserUuid != user.Uuid {
			panic(result.UNAUTHORIZED)
		}

		this.spaceMemberDao.Delete(imageCache)
	}

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
