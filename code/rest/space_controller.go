package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
)

type SpaceController struct {
	BaseController
	spaceDao      *SpaceDao
	matterDao     *MatterDao
	matterService *MatterService
	spaceService  *SpaceService
}

func (this *SpaceController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
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

}

func (this *SpaceController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/space/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/space/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/space/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/space/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

func (this *SpaceController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	//TODO:
	return this.Success("OK")
}

func (this *SpaceController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	space := this.spaceDao.FindByUuid(uuid)

	if space != nil {

		this.spaceDao.Delete(space)
	}

	return this.Success(nil)
}

func (this *SpaceController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	space := this.spaceDao.CheckByUuid(uuid)

	user := this.checkUser(request)

	if space.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(space)

}

func (this *SpaceController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

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

	pager := this.spaceDao.Page(page, pageSize, user.Uuid, sortArray)

	return this.Success(pager)
}
