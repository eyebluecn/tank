package rest

import (
	"net/http"
	"regexp"
	"strconv"

	"time"
)

type UserController struct {
	BaseController
}

//初始化方法
func (this *UserController) Init() {
	this.BaseController.Init()
}

//注册自己的路由。
func (this *UserController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	//每个Controller需要主动注册自己的路由。
	routeMap["/api/user/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/edit"] = this.Wrap(this.Edit, USER_ROLE_USER)
	routeMap["/api/user/change/password"] = this.Wrap(this.ChangePassword, USER_ROLE_USER)
	routeMap["/api/user/reset/password"] = this.Wrap(this.ResetPassword, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/login"] = this.Wrap(this.Login, USER_ROLE_GUEST)
	routeMap["/api/user/logout"] = this.Wrap(this.Logout, USER_ROLE_GUEST)
	routeMap["/api/user/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/user/page"] = this.Wrap(this.Page, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/disable"] = this.Wrap(this.Disable, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/enable"] = this.Wrap(this.Enable, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

//使用用户名和密码进行登录。
//参数：
// @email:邮箱
// @password:密码
func (this *UserController) Login(writer http.ResponseWriter, request *http.Request) *WebResult {

	email := request.FormValue("email")
	password := request.FormValue("password")

	if "" == email || "" == password {

		this.PanicBadRequest("请输入邮箱和密码")
	}

	user := this.userDao.FindByEmail(email)
	if user == nil {

		this.PanicBadRequest("邮箱或密码错误")
	} else {
		if !MatchBcrypt(password, user.Password) {

			this.PanicBadRequest("邮箱或密码错误")
		}
	}

	//登录成功，设置Cookie。有效期30天。
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 30)

	//持久化用户的session.
	session := &Session{
		UserUuid:   user.Uuid,
		Ip:         GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	//设置用户的cookie.
	cookie := http.Cookie{
		Name:    COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   session.Uuid,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	//更新用户上次登录时间和ip
	user.LastTime = time.Now()
	user.LastIp = GetIpAddress(request)
	this.userDao.Save(user)

	return this.Success(user)
}

//创建一个用户
func (this *UserController) Create(writer http.ResponseWriter, request *http.Request) *WebResult {

	username := request.FormValue("username")
	if m, _ := regexp.MatchString(`^[0-9a-zA-Z_]+$`, username); !m {
		panic(`用户名必填，且只能包含字母，数字和'_''`)
	}
	password := request.FormValue("password")
	if len(password) < 6 {
		panic(`密码长度至少为6位`)
	}

	email := request.FormValue("email")
	if email == "" {
		panic("邮箱必填！")
	}

	avatarUrl := request.FormValue("avatarUrl")
	phone := request.FormValue("phone")
	gender := request.FormValue("gender")
	role := request.FormValue("role")
	city := request.FormValue("city")

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

	//判断重名。
	if this.userDao.CountByUsername(username) > 0 {
		panic(username + "已经被其他用户占用。")
	}
	//判断邮箱重名
	if this.userDao.CountByEmail(email) > 0 {
		panic(email + "已经被其他用户占用。")
	}

	user := &User{
		Role:      GetRole(role),
		Username:  username,
		Password:  GetBcrypt(password),
		Email:     email,
		Phone:     phone,
		Gender:    gender,
		City:      city,
		AvatarUrl: avatarUrl,
		SizeLimit: sizeLimit,
		Status:    USER_STATUS_OK,
	}

	user = this.userDao.Create(user)

	return this.Success(user)
}

//编辑一个用户的资料。
func (this *UserController) Edit(writer http.ResponseWriter, request *http.Request) *WebResult {

	avatarUrl := request.FormValue("avatarUrl")
	uuid := request.FormValue("uuid")
	phone := request.FormValue("phone")
	gender := request.FormValue("gender")
	city := request.FormValue("city")

	currentUser := this.checkUser(writer, request)
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
			this.PanicUnauthorized("没有权限")
		}
	}

	user.AvatarUrl = avatarUrl
	user.Phone = phone
	user.Gender = GetGender(gender)
	user.City = city

	user = this.userDao.Save(user)

	return this.Success(user)
}

//获取用户详情
func (this *UserController) Detail(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)

	return this.Success(user)

}

//退出登录
func (this *UserController) Logout(writer http.ResponseWriter, request *http.Request) *WebResult {

	//session置为过期
	sessionCookie, err := request.Cookie(COOKIE_AUTH_KEY)
	if err != nil {
		return this.Success("已经退出登录了！")
	}
	sessionId := sessionCookie.Value

	user := this.findUser(writer, request)
	if user != nil {
		session := this.sessionDao.FindByUuid(sessionId)
		session.ExpireTime = time.Now()
		this.sessionDao.Save(session)
	}

	//删掉session缓存
	_, err = CONTEXT.SessionCache.Delete(sessionId)
	if err != nil {
		this.logger.Error("删除用户session缓存时出错")
	}

	//清空客户端的cookie.
	expiration := time.Now()
	expiration = expiration.AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:    COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   sessionId,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	return this.Success("退出成功！")
}

//获取用户列表 管理员的权限。
func (this *UserController) Page(writer http.ResponseWriter, request *http.Request) *WebResult {

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

	sortArray := []OrderPair{
		{
			key:   "create_time",
			value: orderCreateTime,
		},
		{
			key:   "update_time",
			value: orderUpdateTime,
		},
		{
			key:   "sort",
			value: orderSort,
		},
		{
			key:   "last_time",
			value: orderLastTime,
		},
	}

	pager := this.userDao.Page(page, pageSize, username, email, phone, status, sortArray)

	return this.Success(pager)
}

//禁用用户
func (this *UserController) Disable(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)

	loginUser := this.checkUser(writer, request)
	if uuid == loginUser.Uuid {
		this.PanicBadRequest("你不能操作自己的状态。")
	}

	if user.Status == USER_STATUS_DISABLED {
		this.PanicBadRequest("用户已经被禁用，操作无效。")
	}

	user.Status = USER_STATUS_DISABLED

	user = this.userDao.Save(user)

	return this.Success(user)

}

//启用用户
func (this *UserController) Enable(writer http.ResponseWriter, request *http.Request) *WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)
	loginUser := this.checkUser(writer, request)
	if uuid == loginUser.Uuid {
		this.PanicBadRequest("你不能操作自己的状态。")
	}

	if user.Status == USER_STATUS_OK {
		this.PanicBadRequest("用户已经是正常状态，操作无效。")
	}

	user.Status = USER_STATUS_OK

	user = this.userDao.Save(user)

	return this.Success(user)

}

//用户修改密码
func (this *UserController) ChangePassword(writer http.ResponseWriter, request *http.Request) *WebResult {

	oldPassword := request.FormValue("oldPassword")
	newPassword := request.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		this.PanicBadRequest("旧密码和新密码都不能为空")
	}

	user := this.checkUser(writer, request)

	//如果是demo账号，不提供修改密码的功能。
	if user.Username == "demo" {
		return this.Success(user)
	}

	if !MatchBcrypt(oldPassword, user.Password) {
		this.PanicBadRequest("旧密码不正确！")
	}

	user.Password = GetBcrypt(newPassword)

	user = this.userDao.Save(user)

	return this.Success(user)
}

//管理员重置用户密码
func (this *UserController) ResetPassword(writer http.ResponseWriter, request *http.Request) *WebResult {

	userUuid := request.FormValue("userUuid")
	password := request.FormValue("password")
	if userUuid == "" {
		this.PanicBadRequest("用户不能为空")
	}
	if password == "" {
		this.PanicBadRequest("密码不能为空")
	}

	currentUser := this.checkUser(writer, request)

	if currentUser.Role != USER_ROLE_ADMINISTRATOR {
		this.PanicUnauthorized("没有权限")
	}

	user := this.userDao.CheckByUuid(userUuid)

	user.Password = GetBcrypt(password)

	user = this.userDao.Save(user)

	return this.Success(currentUser)
}
