package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"regexp"
	"strconv"
)

type SpaceController struct {
	BaseController
	spaceDao      *SpaceDao
	matterDao     *MatterDao
	matterService *MatterService
	spaceService  *SpaceService
	userService   *UserService
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

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
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
	//space's name
	name := request.FormValue("name")
	sizeLimitStr := request.FormValue("sizeLimit")
	totalSizeLimitStr := request.FormValue("totalSizeLimit")

	if name == "" {
		panic("name is required")
	}

	//only admin can edit user's sizeLimit
	var sizeLimit int64 = 0
	if sizeLimitStr == "" {
		panic("space's limit size is required")
	} else {
		intSizeLimit, err := strconv.Atoi(sizeLimitStr)
		if err != nil {
			this.PanicError(err)
		}
		sizeLimit = int64(intSizeLimit)
	}

	var totalSizeLimit int64 = 0
	if totalSizeLimitStr == "" {
		panic("space's total limit size is required")
	} else {
		intTotalSizeLimit, err := strconv.Atoi(totalSizeLimitStr)
		if err != nil {
			this.PanicError(err)
		}
		totalSizeLimit = int64(intTotalSizeLimit)
	}

	//validation work.
	if m, _ := regexp.MatchString(USERNAME_PATTERN, name); !m {
		panic(result.BadRequestI18n(request, i18n.SpaceNameError))
	}

	if this.userDao.CountByUsername(name) > 0 {
		panic(result.BadRequestI18n(request, i18n.SpaceNameExist, name))
	}

	user := this.userService.CreateUser(request, name, "", USER_ROLE_SPACE, sizeLimit, totalSizeLimit)

	//create related space.
	space := this.spaceService.CreateSpace(user.Uuid)
	space.User = user

	return this.Success(space)
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

	//TODO: user fetch self's sapces.
	var pager *Pager
	if user.Role == USER_ROLE_USER {

	} else if user.Role == USER_ROLE_ADMINISTRATOR {
		pager = this.spaceDao.Page(page, pageSize, sortArray)
	}

	//fill the space's user.
	if pager != nil {
		for _, space := range pager.Data.([]*Space) {
			space.User = this.userDao.FindByUuid(space.UserUuid)
		}
	}

	return this.Success(pager)
}
