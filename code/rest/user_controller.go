package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type UserController struct {
	BaseController
	preferenceService *PreferenceService
}

//初始化方法
func (this *UserController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}
}

//注册自己的路由。
func (this *UserController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/user/login"] = this.Wrap(this.Login, USER_ROLE_GUEST)
	routeMap["/api/user/register"] = this.Wrap(this.Register, USER_ROLE_GUEST)
	routeMap["/api/user/edit"] = this.Wrap(this.Edit, USER_ROLE_USER)
	routeMap["/api/user/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/user/logout"] = this.Wrap(this.Logout, USER_ROLE_GUEST)
	routeMap["/api/user/change/password"] = this.Wrap(this.ChangePassword, USER_ROLE_USER)
	routeMap["/api/user/reset/password"] = this.Wrap(this.ResetPassword, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/page"] = this.Wrap(this.Page, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/toggle/status"] = this.Wrap(this.ToggleStatus, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

//使用用户名和密码进行登录。
//参数：
// @username:用户名
// @password:密码
func (this *UserController) Login(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	if "" == username || "" == password {

		panic(result.BadRequest("请输入用户名和密码"))
	}

	user := this.userDao.FindByUsername(username)
	if user == nil {
		panic(result.BadRequest("用户名或密码错误"))
	}

	if !util.MatchBcrypt(password, user.Password) {

		panic(result.BadRequest("用户名或密码错误"))
	}

	//登录成功，设置Cookie。有效期30天。
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 30)

	//持久化用户的session.
	session := &Session{
		UserUuid:   user.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	//设置用户的cookie.
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   session.Uuid,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	//更新用户上次登录时间和ip
	user.LastTime = time.Now()
	user.LastIp = util.GetIpAddress(request)
	this.userDao.Save(user)

	return this.Success(user)
}

//用户自主注册。
func (this *UserController) Register(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	if m, _ := regexp.MatchString(`^[0-9a-zA-Z_]+$`, username); !m {
		panic(`用户名必填，且只能包含字母，数字和'_''`)
	}

	if len(password) < 6 {
		panic(`密码长度至少为6位`)
	}

	//判断重名。
	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequest("%s已经被其他用户占用。", username))
	}

	preference := this.preferenceService.Fetch()

	user := &User{
		Role:      USER_ROLE_USER,
		Username:  username,
		Password:  util.GetBcrypt(password),
		SizeLimit: preference.DefaultTotalSizeLimit,
		Status:    USER_STATUS_OK,
	}

	user = this.userDao.Create(user)

	return this.Success(user)
}

//编辑一个用户的资料。
func (this *UserController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	avatarUrl := request.FormValue("avatarUrl")
	uuid := request.FormValue("uuid")

	currentUser := this.checkUser(request)
	user := this.userDao.CheckByUuid(uuid)

	if currentUser.Role == USER_ROLE_ADMINISTRATOR {
		//只有管理员可以改变用户上传的大小
		//判断用户上传大小限制。
		sizeLimitStr := request.FormValue("sizeLimit")
		var sizeLimit int64 = 0
		if sizeLimitStr == "" {
			panic("用户上传限制必填！")
		} else {
			intsizeLimit, err := strconv.Atoi(sizeLimitStr)
			if err != nil {
				this.PanicError(err)
			}
			sizeLimit = int64(intsizeLimit)
		}
		user.SizeLimit = sizeLimit
	} else {
		if currentUser.Uuid != uuid {
			panic(result.Unauthorized("没有权限"))
		}
	}

	user.AvatarUrl = avatarUrl

	user = this.userDao.Save(user)

	return this.Success(user)
}

//获取用户详情
func (this *UserController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)

	return this.Success(user)

}

//退出登录
func (this *UserController) Logout(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//session置为过期
	sessionCookie, err := request.Cookie(core.COOKIE_AUTH_KEY)
	if err != nil {
		return this.Success("已经退出登录了！")
	}
	sessionId := sessionCookie.Value

	user := this.findUser(request)
	if user != nil {
		session := this.sessionDao.FindByUuid(sessionId)
		session.ExpireTime = time.Now()
		this.sessionDao.Save(session)
	}

	//删掉session缓存
	_, err = core.CONTEXT.GetSessionCache().Delete(sessionId)
	if err != nil {
		this.logger.Error("删除用户session缓存时出错")
	}

	//清空客户端的cookie.
	expiration := time.Now()
	expiration = expiration.AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   sessionId,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	return this.Success("退出成功！")
}

//获取用户列表 管理员的权限。
func (this *UserController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")

	username := request.FormValue("username")
	email := request.FormValue("email")
	phone := request.FormValue("phone")
	status := request.FormValue("status")
	orderLastTime := request.FormValue("orderLastTime")

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
		{
			Key:   "update_time",
			Value: orderUpdateTime,
		},
		{
			Key:   "sort",
			Value: orderSort,
		},
		{
			Key:   "last_time",
			Value: orderLastTime,
		},
	}

	pager := this.userDao.Page(page, pageSize, username, email, phone, status, sortArray)

	return this.Success(pager)
}

//修改用户状态
func (this *UserController) ToggleStatus(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	user := this.checkUser(request)
	if uuid == user.Uuid {
		panic(result.Unauthorized("你不能操作自己的状态。"))
	}

	if currentUser.Status == USER_STATUS_OK {
		currentUser.Status = USER_STATUS_DISABLED
	} else if currentUser.Status == USER_STATUS_DISABLED {
		currentUser.Status = USER_STATUS_OK
	}

	currentUser = this.userDao.Save(currentUser)

	return this.Success(currentUser)

}

//用户修改密码
func (this *UserController) ChangePassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	oldPassword := request.FormValue("oldPassword")
	newPassword := request.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		panic(result.BadRequest("旧密码和新密码都不能为空"))
	}

	user := this.checkUser(request)

	//如果是demo账号，不提供修改密码的功能。
	if user.Username == "demo" {
		return this.Success(user)
	}

	if !util.MatchBcrypt(oldPassword, user.Password) {
		panic(result.BadRequest("旧密码不正确！"))
	}

	user.Password = util.GetBcrypt(newPassword)

	user = this.userDao.Save(user)

	return this.Success(user)
}

//管理员重置用户密码
func (this *UserController) ResetPassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	userUuid := request.FormValue("userUuid")
	password := request.FormValue("password")
	if userUuid == "" {
		panic(result.BadRequest("用户不能为空"))
	}
	if password == "" {
		panic(result.BadRequest("密码不能为空"))
	}

	currentUser := this.checkUser(request)

	if currentUser.Role != USER_ROLE_ADMINISTRATOR {
		panic(result.Unauthorized("没有权限"))
	}

	user := this.userDao.CheckByUuid(userUuid)

	user.Password = util.GetBcrypt(password)

	user = this.userDao.Save(user)

	return this.Success(currentUser)
}
