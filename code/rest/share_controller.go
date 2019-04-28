package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ShareController struct {
	BaseController
	shareDao     *ShareDao
	bridgeDao    *BridgeDao
	matterDao    *MatterDao
	shareService *ShareService
}

//初始化方法
func (this *ShareController) Init() {
	this.BaseController.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if b, ok := b.(*ShareService); ok {
		this.shareService = b
	}

}

//注册自己的路由。
func (this *ShareController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/share/create"] = this.Wrap(this.Create, USER_ROLE_USER)
	routeMap["/api/share/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/share/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/share/page"] = this.Wrap(this.Page, USER_ROLE_USER)

	return routeMap
}

//删除一条记录
func (this *ShareController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	matterUuids := request.FormValue("matterUuids")
	expireTimeStr := request.FormValue("expireTime")

	if matterUuids == "" {
		panic(result.BadRequest("matterUuids必填"))
	}

	var expireTime time.Time
	if expireTimeStr == "" {
		panic(result.BadRequest("时间格式错误！"))
	} else {
		expireTime = util.ConvertDateTimeStringToTime(expireTimeStr)
	}

	if expireTime.Before(time.Now()) {
		panic(result.BadRequest("过期时间错误！"))
	}

	uuidArray := strings.Split(matterUuids, ",")

	if len(uuidArray) == 0 {
		panic(result.BadRequest("请至少分享一个文件"))
	}

	user := this.checkUser(writer, request)
	for _, uuid := range uuidArray {

		matter := this.matterDao.CheckByUuid(uuid)

		//判断文件的所属人是否正确
		if matter.UserUuid != user.Uuid {
			panic(result.Unauthorized("没有权限"))
		}
	}

	//创建share记录
	share := &Share{
		UserUuid:      user.Uuid,
		DownloadTimes: 0,
		Code:          util.RandomString4(),
		ExpireTime:    expireTime,
	}
	this.shareDao.Create(share)

	//创建关联的matter
	for _, matterUuid := range uuidArray {
		bridge := &Bridge{
			ShareUuid:  share.Uuid,
			MatterUuid: matterUuid,
		}
		this.bridgeDao.Create(bridge)
	}

	return this.Success(share)
}

//删除一条记录
func (this *ShareController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid必填"))
	}

	share := this.shareDao.FindByUuid(uuid)

	if share != nil {
		this.shareDao.Delete(share)
	}

	return this.Success(nil)
}

//查看详情。
func (this *ShareController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("分享的uuid必填"))
	}

	share := this.shareDao.CheckByUuid(uuid)

	//验证当前之人是否有权限查看这么详细。
	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		if share.UserUuid != user.Uuid {
			panic(result.Unauthorized("没有权限"))
		}
	}

	return this.Success(share)

}

//按照分页的方式查询
func (this *ShareController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//如果是根目录，那么就传入root.
	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	userUuid := request.FormValue("userUuid")
	orderCreateTime := request.FormValue("orderCreateTime")

	user := this.checkUser(writer, request)
	if user.Role != USER_ROLE_ADMINISTRATOR {
		userUuid = user.Uuid
	}

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

	pager := this.shareDao.Page(page, pageSize, userUuid, sortArray)

	return this.Success(pager)
}
