package result

import (
	"fmt"
	"net/http"
)

type WebResult struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (this *WebResult) Error() string {
	return this.Msg
}

type CodeWrapper struct {
	Code        string
	HttpStatus  int
	Description string
}

var (
	OK                      = &CodeWrapper{Code: "OK", HttpStatus: http.StatusOK, Description: "成功"}
	BAD_REQUEST             = &CodeWrapper{Code: "BAD_REQUEST", HttpStatus: http.StatusBadRequest, Description: "请求不合法"}
	CAPTCHA_ERROR           = &CodeWrapper{Code: "CAPTCHA_ERROR", HttpStatus: http.StatusBadRequest, Description: "验证码错误"}
	NEED_CAPTCHA            = &CodeWrapper{Code: "NEED_CAPTCHA", HttpStatus: http.StatusBadRequest, Description: "验证码必填"}
	NEED_SHARE_CODE         = &CodeWrapper{Code: "NEED_SHARE_CODE", HttpStatus: http.StatusUnauthorized, Description: "分享提取码必填"}
	SHARE_CODE_ERROR        = &CodeWrapper{Code: "SHARE_CODE_ERROR", HttpStatus: http.StatusUnauthorized, Description: "分享提取码错误"}
	USERNAME_PASSWORD_ERROR = &CodeWrapper{Code: "USERNAME_PASSWORD_ERROR", HttpStatus: http.StatusBadRequest, Description: "用户名或密码错误"}
	PARAMS_ERROR            = &CodeWrapper{Code: "PARAMS_ERROR", HttpStatus: http.StatusBadRequest, Description: "用户名或密码错误"}
	LOGIN                   = &CodeWrapper{Code: "LOGIN", HttpStatus: http.StatusUnauthorized, Description: "未登录，禁止访问"}
	LOGIN_EXPIRE            = &CodeWrapper{Code: "LOGIN_EXPIRE", HttpStatus: http.StatusUnauthorized, Description: "登录过期，请重新登录"}
	USER_DISABLED           = &CodeWrapper{Code: "USER_DISABLED", HttpStatus: http.StatusForbidden, Description: "账户被禁用，禁止访问"}
	UNAUTHORIZED            = &CodeWrapper{Code: "UNAUTHORIZED", HttpStatus: http.StatusUnauthorized, Description: "没有权限，禁止访问"}
	NOT_FOUND               = &CodeWrapper{Code: "NOT_FOUND", HttpStatus: http.StatusNotFound, Description: "内容不存在"}
	RANGE_NOT_SATISFIABLE   = &CodeWrapper{Code: "RANGE_NOT_SATISFIABLE", HttpStatus: http.StatusRequestedRangeNotSatisfiable, Description: "文件范围读取错误"}
	NOT_INSTALLED           = &CodeWrapper{Code: "NOT_INSTALLED", HttpStatus: http.StatusInternalServerError, Description: "系统尚未安装"}
	SERVER                  = &CodeWrapper{Code: "SERVER", HttpStatus: http.StatusInternalServerError, Description: "服务器出错"}
	UNKNOWN                 = &CodeWrapper{Code: "UNKNOWN", HttpStatus: http.StatusInternalServerError, Description: "服务器未知错误"}
)

//根据 CodeWrapper来获取对应的HttpStatus
func FetchHttpStatus(code string) int {
	if code == OK.Code {
		return OK.HttpStatus
	} else if code == BAD_REQUEST.Code {
		return BAD_REQUEST.HttpStatus
	} else if code == CAPTCHA_ERROR.Code {
		return CAPTCHA_ERROR.HttpStatus
	} else if code == NEED_CAPTCHA.Code {
		return NEED_CAPTCHA.HttpStatus
	} else if code == NEED_SHARE_CODE.Code {
		return NEED_SHARE_CODE.HttpStatus
	} else if code == SHARE_CODE_ERROR.Code {
		return SHARE_CODE_ERROR.HttpStatus
	} else if code == USERNAME_PASSWORD_ERROR.Code {
		return USERNAME_PASSWORD_ERROR.HttpStatus
	} else if code == PARAMS_ERROR.Code {
		return PARAMS_ERROR.HttpStatus
	} else if code == LOGIN.Code {
		return LOGIN.HttpStatus
	} else if code == LOGIN_EXPIRE.Code {
		return LOGIN_EXPIRE.HttpStatus
	} else if code == USER_DISABLED.Code {
		return USER_DISABLED.HttpStatus
	} else if code == UNAUTHORIZED.Code {
		return UNAUTHORIZED.HttpStatus
	} else if code == NOT_FOUND.Code {
		return NOT_FOUND.HttpStatus
	} else if code == RANGE_NOT_SATISFIABLE.Code {
		return RANGE_NOT_SATISFIABLE.HttpStatus
	} else if code == NOT_INSTALLED.Code {
		return NOT_INSTALLED.HttpStatus
	} else if code == SERVER.Code {
		return SERVER.HttpStatus
	} else {
		return UNKNOWN.HttpStatus
	}
}

func ConstWebResult(codeWrapper *CodeWrapper) *WebResult {

	wr := &WebResult{
		Code: codeWrapper.Code,
		Msg:  codeWrapper.Description,
	}
	return wr
}

func CustomWebResult(codeWrapper *CodeWrapper, description string) *WebResult {

	if description == "" {
		description = codeWrapper.Description
	}
	wr := &WebResult{
		Code: codeWrapper.Code,
		Msg:  description,
	}
	return wr
}

//请求参数有问题
func BadRequest(format string, v ...interface{}) *WebResult {
	return CustomWebResult(BAD_REQUEST, fmt.Sprintf(format, v...))
}

//没有权限
func Unauthorized(format string, v ...interface{}) *WebResult {
	return CustomWebResult(UNAUTHORIZED, fmt.Sprintf(format, v...))
}

//没有找到
func NotFound(format string, v ...interface{}) *WebResult {
	return CustomWebResult(NOT_FOUND, fmt.Sprintf(format, v...))

}

//服务器内部出问题
func Server(format string, v ...interface{}) *WebResult {
	return CustomWebResult(SERVER, fmt.Sprintf(format, v...))
}

//所有的数据库错误情况
var (
	DB_ERROR_DUPLICATE_KEY  = "Error 1062: Duplicate entry"
	DB_ERROR_NOT_FOUND      = "record not found"
	DB_TOO_MANY_CONNECTIONS = "Error 1040: Too many connections"
	DB_BAD_CONNECTION       = "driver: bad connection"
)
