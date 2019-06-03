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
	userService       *UserService
}

func (this *UserController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
	}

}

func (this *UserController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/user/login"] = this.Wrap(this.Login, USER_ROLE_GUEST)
	routeMap["/api/user/authentication/login"] = this.Wrap(this.AuthenticationLogin, USER_ROLE_GUEST)
	routeMap["/api/user/register"] = this.Wrap(this.Register, USER_ROLE_GUEST)
	routeMap["/api/user/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
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

	if user.Status == USER_STATUS_DISABLED {
		panic(result.BadRequestI18n(request, i18n.UserDisabled))
	}

	//set cookie. expire after 30 days.
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 30)

	//save session to db.
	session := &Session{
		UserUuid:   user.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	//set cookie
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   session.Uuid,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	//update lastTime and lastIp
	user.LastTime = time.Now()
	user.LastIp = util.GetIpAddress(request)
	this.userDao.Save(user)
}

//login by username and password
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

//login by authentication.
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

//register by username and password. After registering, will auto login.
func (this *UserController) Register(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	preference := this.preferenceService.Fetch()
	if !preference.AllowRegister {
		panic(result.BadRequestI18n(request, i18n.UserRegisterNotAllowd))
	}

	if m, _ := regexp.MatchString(USERNAME_PATTERN, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

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

	//auto login
	this.innerLogin(writer, request, user)

	return this.Success(user)
}

func (this *UserController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")
	role := request.FormValue("role")
	sizeLimitStr := request.FormValue("sizeLimit")
	totalSizeLimitStr := request.FormValue("totalSizeLimit")

	//only admin can edit user's sizeLimit
	var sizeLimit int64 = 0
	if sizeLimitStr == "" {
		panic("user's limit size is required")
	} else {
		intSizeLimit, err := strconv.Atoi(sizeLimitStr)
		if err != nil {
			this.PanicError(err)
		}
		sizeLimit = int64(intSizeLimit)
	}

	var totalSizeLimit int64 = 0
	if totalSizeLimitStr == "" {
		panic("user's total limit size is required")
	} else {
		intTotalSizeLimit, err := strconv.Atoi(totalSizeLimitStr)
		if err != nil {
			this.PanicError(err)
		}
		totalSizeLimit = int64(intTotalSizeLimit)
	}

	if m, _ := regexp.MatchString(USERNAME_PATTERN, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, username))
	}

	if role != USER_ROLE_USER && role != USER_ROLE_ADMINISTRATOR {
		role = USER_ROLE_USER
	}

	user := &User{
		Username:       username,
		Password:       util.GetBcrypt(password),
		Role:           role,
		SizeLimit:      sizeLimit,
		TotalSizeLimit: totalSizeLimit,
		Status:         USER_STATUS_OK,
	}

	user = this.userDao.Create(user)

	return this.Success(user)
}

func (this *UserController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	avatarUrl := request.FormValue("avatarUrl")
	sizeLimitStr := request.FormValue("sizeLimit")
	totalSizeLimitStr := request.FormValue("totalSizeLimit")
	role := request.FormValue("role")

	user := this.checkUser(request)
	currentUser := this.userDao.CheckByUuid(uuid)

	currentUser.AvatarUrl = avatarUrl

	if user.Role == USER_ROLE_ADMINISTRATOR {
		//only admin can edit user's sizeLimit
		var sizeLimit int64 = 0
		if sizeLimitStr == "" {
			panic("user's limit size is required")
		} else {
			intSizeLimit, err := strconv.Atoi(sizeLimitStr)
			if err != nil {
				this.PanicError(err)
			}
			sizeLimit = int64(intSizeLimit)
		}
		currentUser.SizeLimit = sizeLimit

		var totalSizeLimit int64 = 0
		if totalSizeLimitStr == "" {
			panic("user's total limit size is required")
		} else {
			intTotalSizeLimit, err := strconv.Atoi(totalSizeLimitStr)
			if err != nil {
				this.PanicError(err)
			}
			totalSizeLimit = int64(intTotalSizeLimit)
		}
		currentUser.TotalSizeLimit = totalSizeLimit

		if role == USER_ROLE_USER || role == USER_ROLE_ADMINISTRATOR {
			currentUser.Role = role
		}

	} else if user.Uuid != uuid {
		panic(result.UNAUTHORIZED)
	}

	currentUser = this.userDao.Save(currentUser)

	return this.Success(currentUser)
}

func (this *UserController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)

	return this.Success(user)

}

func (this *UserController) Logout(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//try to find from SessionCache.
	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)
	if sessionId == "" {
		return nil
	}

	user := this.findUser(request)
	if user != nil {
		session := this.sessionDao.FindByUuid(sessionId)
		session.ExpireTime = time.Now()
		this.sessionDao.Save(session)
	}

	//delete session.
	_, err := core.CONTEXT.GetSessionCache().Delete(sessionId)
	if err != nil {
		this.logger.Error("error while deleting session.")
	}

	//clear cookie.
	expiration := time.Now()
	expiration = expiration.AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   sessionId,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	return this.Success("OK")
}

func (this *UserController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")

	username := request.FormValue("username")
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

	pager := this.userDao.Page(page, pageSize, username, status, sortArray)

	return this.Success(pager)
}

func (this *UserController) ToggleStatus(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	user := this.checkUser(request)
	if uuid == user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	if currentUser.Status == USER_STATUS_OK {
		currentUser.Status = USER_STATUS_DISABLED
	} else if currentUser.Status == USER_STATUS_DISABLED {
		currentUser.Status = USER_STATUS_OK
	}

	currentUser = this.userDao.Save(currentUser)

	cacheUsers := this.userService.FindCacheUsersByUuid(currentUser.Uuid)
	this.logger.Info("find %d cache users", len(cacheUsers))
	for _, u := range cacheUsers {
		u.Status = currentUser.Status
	}

	return this.Success(currentUser)

}

func (this *UserController) Transfiguration(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)

	//expire after 10 minutes.
	expiration := time.Now()
	expiration = expiration.Add(10 * time.Minute)

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

func (this *UserController) ChangePassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	oldPassword := request.FormValue("oldPassword")
	newPassword := request.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		panic(result.BadRequest("oldPassword and newPassword cannot be null"))
	}

	user := this.checkUser(request)

	//if username is demo, cannot change password.
	if user.Username == USERNAME_DEMO {
		return this.Success(user)
	}

	if !util.MatchBcrypt(oldPassword, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UserOldPasswordError))
	}

	user.Password = util.GetBcrypt(newPassword)

	user = this.userDao.Save(user)

	return this.Success(user)
}

//admin reset password.
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
