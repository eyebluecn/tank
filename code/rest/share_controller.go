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
	shareDao      *ShareDao
	bridgeDao     *BridgeDao
	matterDao     *MatterDao
	matterService *MatterService
	shareService  *ShareService
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

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
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
	routeMap["/api/share/delete/batch"] = this.Wrap(this.DeleteBatch, USER_ROLE_USER)
	routeMap["/api/share/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/share/page"] = this.Wrap(this.Page, USER_ROLE_USER)
	routeMap["/api/share/browse"] = this.Wrap(this.Browse, USER_ROLE_GUEST)

	return routeMap
}

//创建一个分享
func (this *ShareController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	matterUuids := request.FormValue("matterUuids")
	expireInfinityStr := request.FormValue("expireInfinity")
	expireTimeStr := request.FormValue("expireTime")

	if matterUuids == "" {
		panic(result.BadRequest("matterUuids必填"))
	}

	var expireTime time.Time
	expireInfinity := false
	if expireInfinityStr == TRUE {
		expireInfinity = true
		expireTime = time.Now()
	} else {

		if expireTimeStr == "" {
			panic(result.BadRequest("时间格式错误！"))
		} else {
			expireTime = util.ConvertDateTimeStringToTime(expireTimeStr)
		}

		if expireTime.Before(time.Now()) {
			panic(result.BadRequest("过期时间错误！"))
		}

	}

	uuidArray := strings.Split(matterUuids, ",")

	if len(uuidArray) == 0 {
		panic(result.BadRequest("请至少分享一个文件"))
	} else if len(uuidArray) > SHARE_MAX_NUM {
		panic(result.BadRequest("一次分享文件数不能超过 %d", SHARE_MAX_NUM))
	}

	var name string
	shareType := SHARE_TYPE_MIX
	user := this.checkUser(writer, request)
	var puuid string
	var matters []*Matter
	for key, uuid := range uuidArray {

		matter := this.matterDao.CheckByUuid(uuid)

		//判断文件的所属人是否正确
		if matter.UserUuid != user.Uuid {
			panic(result.Unauthorized("不是你的文件，没有权限"))
		}

		matters = append(matters, matter)

		if key == 0 {
			puuid = matter.Puuid
			name = matter.Name
			if matter.Dir {
				shareType = SHARE_TYPE_DIRECTORY
			} else {
				shareType = SHARE_TYPE_FILE
			}
		} else {
			if matter.Puuid != puuid {
				panic(result.Unauthorized("一次只能分享同一个文件夹中的内容"))
			}
		}

	}

	if len(matters) > 1 {
		shareType = SHARE_TYPE_MIX
		name = matters[0].Name + "," + matters[1].Name + " 等"
	}

	//创建share记录
	share := &Share{
		Name:           name,
		ShareType:      shareType,
		UserUuid:       user.Uuid,
		Username:       user.Username,
		DownloadTimes:  0,
		Code:           util.RandomString4(),
		ExpireInfinity: expireInfinity,
		ExpireTime:     expireTime,
	}
	this.shareDao.Create(share)

	//创建关联的matter
	for _, matter := range matters {
		bridge := &Bridge{
			ShareUuid:  share.Uuid,
			MatterUuid: matter.Uuid,
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

		//删除对应的bridge.
		this.bridgeDao.DeleteByShareUuid(share.Uuid)

		this.shareDao.Delete(share)
	}

	return this.Success(nil)
}

//删除一系列分享
func (this *ShareController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("uuids必填"))
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		imageCache := this.shareDao.FindByUuid(uuid)

		//判断图片缓存的所属人是否正确
		user := this.checkUser(writer, request)
		if user.Role != USER_ROLE_ADMINISTRATOR && imageCache.UserUuid != user.Uuid {
			panic(result.Unauthorized("没有权限"))
		}

		this.shareDao.Delete(imageCache)
	}

	return this.Success("删除成功！")
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

//验证提取码对应的某个shareUuid是否有效
func (this *ShareController) CheckShare(writer http.ResponseWriter, request *http.Request) *Share {

	//如果是根目录，那么就传入root.
	shareUuid := request.FormValue("shareUuid")
	code := request.FormValue("code")
	user := this.findUser(writer, request)

	return this.shareService.CheckShare(shareUuid, code, user)
}

//浏览某个分享中的文件
func (this *ShareController) Browse(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//如果是根目录，那么就传入root.
	shareUuid := request.FormValue("shareUuid")
	code := request.FormValue("code")

	//当前查看的puuid。 puuid=root表示查看分享的根目录，其余表示查看某个文件夹下的文件。
	puuid := request.FormValue("puuid")
	rootUuid := request.FormValue("rootUuid")

	user := this.findUser(writer, request)
	share := this.shareService.CheckShare(shareUuid, code, user)
	bridges := this.bridgeDao.ListByShareUuid(share.Uuid)

	if puuid == "" {
		puuid = MATTER_ROOT
	}
	//分享的跟目录
	if puuid == MATTER_ROOT {

		//获取对应的 matter.
		var matters []*Matter
		if len(bridges) != 0 {
			uuids := make([]string, 0)
			for _, bridge := range bridges {
				uuids = append(uuids, bridge.MatterUuid)
			}

			sortArray := []builder.OrderPair{
				{
					Key:   "dir",
					Value: DIRECTION_DESC,
				},
			}
			matters = this.matterDao.ListByUuids(uuids, sortArray)

			share.Matters = matters
		}

	} else {

		//如果当前查看的目录就是根目录，那么无需再验证
		if puuid == rootUuid {
			dirMatter := this.matterDao.CheckByUuid(puuid)
			share.DirMatter = dirMatter
		} else {
			dirMatter := this.matterService.Detail(puuid)

			//验证 shareRootMatter是否在被分享。
			shareRootMatter := this.matterDao.CheckByUuid(rootUuid)
			if !shareRootMatter.Dir {
				panic(result.BadRequest("只有文件夹可以浏览！"))
			}
			this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, shareRootMatter.Uuid)

			//到rootUuid的地方掐断。
			find := false
			parentMatter := dirMatter.Parent
			for parentMatter != nil {
				if parentMatter.Uuid == rootUuid {
					parentMatter.Parent = nil
					find = true
					break
				}
				parentMatter = parentMatter.Parent
			}

			if !find {
				panic(result.BadRequest("rootUuid不是分享的根目录"))
			}

			share.DirMatter = dirMatter
		}

	}

	return this.Success(share)

}
