package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
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

func (this *UserController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}
}

func (this *UserController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/user/login"] = this.Wrap(this.Login, USER_ROLE_GUEST)
	routeMap["/api/user/authentication/login"] = this.Wrap(this.AuthenticationLogin, USER_ROLE_GUEST)
	routeMap["/api/user/register"] = this.Wrap(this.Register, USER_ROLE_GUEST)
	routeMap["/api/user/edit"] = this.Wrap(this.Edit, USER_ROLE_USER)
	routeMap["/api/user/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/user/logout"] = this.Wrap(this.Logout, USER_ROLE_GUEST)
	routeMap["/api/user/change/password"] = this.Wrap(this.ChangePassword, USER_ROLE_USER)
	routeMap["/api/user/reset/password"] = this.Wrap(this.ResetPassword, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/page"] = this.Wrap(this.Page, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/toggle/status"] = this.Wrap(this.ToggleStatus, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/transfiguration"] = this.Wrap(this.Transfiguration, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

func (this *UserController) innerLogin(writer http.ResponseWriter, request *http.Request, user *User) {

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
}

//使用用户名和密码进行登录。
//参数：
// @username:用户名
// @password:密码
func (this *UserController) Login(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	if "" == username || "" == password {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordCannotNull))
	}

	user := this.userDao.FindByUsername(username)
	if user == nil {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	}

	if !util.MatchBcrypt(password, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	}
	this.innerLogin(writer, request, user)

	return this.Success(user)
}

//使用Authentication进行登录。
func (this *UserController) AuthenticationLogin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	authentication := request.FormValue("authentication")
	if authentication == "" {
		panic(result.BadRequest("authentication cannot be null"))
	}
	session := this.sessionDao.FindByUuid(authentication)
	if session == nil {
		panic(result.BadRequest("authentication error"))
	}
	duration := session.ExpireTime.Sub(time.Now())
	if duration <= 0 {
		panic(result.BadRequest("login info has expired"))
	}

	user := this.userDao.CheckByUuid(session.UserUuid)
	this.innerLogin(writer, request, user)
	return this.Success(user)
}

//用户自主注册。
func (this *UserController) Register(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	preference := this.preferenceService.Fetch()
	if !preference.AllowRegister {
		panic(result.BadRequestI18n(request, i18n.UserRegisterNotAllowd))
	}

	if m, _ := regexp.MatchString(`^[0-9a-zA-Z_]+$`, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

	//判断重名。
	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, username))
	}

	user := &User{
		Role:      USER_ROLE_USER,
		Username:  username,
		Password:  util.GetBcrypt(password),
		SizeLimit: preference.DefaultTotalSizeLimit,
		Status:    USER_STATUS_OK,
	}

	user = this.userDao.Create(user)

	//做一次登录操作
	this.innerLogin(writer, request, user)

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
			panic(result.UNAUTHORIZED)
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

//变身为指定用户。
func (this *UserController) Transfiguration(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)

	//有效期10分钟
	expiration := time.Now()
	expiration = expiration.Add(10 * time.Minute)

	//持久化用户的session.
	session := &Session{
		UserUuid:   currentUser.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	return this.Success(session.Uuid)
}

//用户修改密码
func (this *UserController) ChangePassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	oldPassword := request.FormValue("oldPassword")
	newPassword := request.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		panic(result.BadRequest("oldPassword and newPassword cannot be null"))
	}

	user := this.checkUser(request)

	//如果是demo账号，不提供修改密码的功能。
	if user.Username == "demo" {
		return this.Success(user)
	}

	if !util.MatchBcrypt(oldPassword, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UserOldPasswordError))
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
		panic(result.BadRequest("userUuid cannot be null"))
	}
	if password == "" {
		panic(result.BadRequest("password cannot be null"))
	}

	currentUser := this.checkUser(request)

	if currentUser.Role != USER_ROLE_ADMINISTRATOR {
		panic(result.UNAUTHORIZED)
	}

	user := this.userDao.CheckByUuid(userUuid)

	user.Password = util.GetBcrypt(password)

	user = this.userDao.Save(user)

	return this.Success(currentUser)
}
