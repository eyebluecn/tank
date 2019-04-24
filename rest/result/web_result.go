package result

import "net/http"

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
	CODE_WRAPPER_OK                      = &CodeWrapper{Code: "OK", HttpStatus: http.StatusOK, Description: "成功"}
	CODE_WRAPPER_BAD_REQUEST             = &CodeWrapper{Code: "BAD_REQUEST", HttpStatus: http.StatusBadRequest, Description: "请求不合法"}
	CODE_WRAPPER_CAPTCHA_ERROR           = &CodeWrapper{Code: "CAPTCHA_ERROR", HttpStatus: http.StatusBadRequest, Description: "验证码错误"}
	CODE_WRAPPER_NEED_CAPTCHA            = &CodeWrapper{Code: "NEED_CAPTCHA", HttpStatus: http.StatusBadRequest, Description: "验证码必填"}
	CODE_WRAPPER_USERNAME_PASSWORD_ERROR = &CodeWrapper{Code: "USERNAME_PASSWORD_ERROR", HttpStatus: http.StatusBadRequest, Description: "用户名或密码错误"}
	CODE_WRAPPER_PARAMS_ERROR            = &CodeWrapper{Code: "PARAMS_ERROR", HttpStatus: http.StatusBadRequest, Description: "用户名或密码错误"}
	CODE_WRAPPER_LOGIN                   = &CodeWrapper{Code: "LOGIN", HttpStatus: http.StatusUnauthorized, Description: "未登录，禁止访问"}
	CODE_WRAPPER_LOGIN_EXPIRE            = &CodeWrapper{Code: "LOGIN_EXPIRE", HttpStatus: http.StatusUnauthorized, Description: "登录过期，请重新登录"}
	CODE_WRAPPER_USER_DISABLED           = &CodeWrapper{Code: "USER_DISABLED", HttpStatus: http.StatusForbidden, Description: "账户被禁用，禁止访问"}
	CODE_WRAPPER_UNAUTHORIZED            = &CodeWrapper{Code: "UNAUTHORIZED", HttpStatus: http.StatusUnauthorized, Description: "没有权限，禁止访问"}
	CODE_WRAPPER_NOT_FOUND               = &CodeWrapper{Code: "NOT_FOUND", HttpStatus: http.StatusNotFound, Description: "内容不存在"}
	CODE_WRAPPER_RANGE_NOT_SATISFIABLE   = &CodeWrapper{Code: "RANGE_NOT_SATISFIABLE", HttpStatus: http.StatusRequestedRangeNotSatisfiable, Description: "文件范围读取错误"}
	CODE_WRAPPER_NOT_INSTALLED           = &CodeWrapper{Code: "NOT_INSTALLED", HttpStatus: http.StatusInternalServerError, Description: "系统尚未安装"}
	CODE_WRAPPER_SERVER                  = &CodeWrapper{Code: "SERVER", HttpStatus: http.StatusInternalServerError, Description: "服务器出错"}
	CODE_WRAPPER_UNKNOWN                 = &CodeWrapper{Code: "UNKNOWN", HttpStatus: http.StatusInternalServerError, Description: "服务器未知错误"}
)

//根据 CodeWrapper来获取对应的HttpStatus
func FetchHttpStatus(code string) int {
	if code == CODE_WRAPPER_OK.Code {
		return CODE_WRAPPER_OK.HttpStatus
	} else if code == CODE_WRAPPER_BAD_REQUEST.Code {
		return CODE_WRAPPER_BAD_REQUEST.HttpStatus
	} else if code == CODE_WRAPPER_CAPTCHA_ERROR.Code {
		return CODE_WRAPPER_CAPTCHA_ERROR.HttpStatus
	} else if code == CODE_WRAPPER_NEED_CAPTCHA.Code {
		return CODE_WRAPPER_NEED_CAPTCHA.HttpStatus
	} else if code == CODE_WRAPPER_USERNAME_PASSWORD_ERROR.Code {
		return CODE_WRAPPER_USERNAME_PASSWORD_ERROR.HttpStatus
	} else if code == CODE_WRAPPER_PARAMS_ERROR.Code {
		return CODE_WRAPPER_PARAMS_ERROR.HttpStatus
	} else if code == CODE_WRAPPER_LOGIN.Code {
		return CODE_WRAPPER_LOGIN.HttpStatus
	} else if code == CODE_WRAPPER_LOGIN_EXPIRE.Code {
		return CODE_WRAPPER_LOGIN_EXPIRE.HttpStatus
	} else if code == CODE_WRAPPER_USER_DISABLED.Code {
		return CODE_WRAPPER_USER_DISABLED.HttpStatus
	} else if code == CODE_WRAPPER_UNAUTHORIZED.Code {
		return CODE_WRAPPER_UNAUTHORIZED.HttpStatus
	} else if code == CODE_WRAPPER_NOT_FOUND.Code {
		return CODE_WRAPPER_NOT_FOUND.HttpStatus
	} else if code == CODE_WRAPPER_RANGE_NOT_SATISFIABLE.Code {
		return CODE_WRAPPER_RANGE_NOT_SATISFIABLE.HttpStatus
	} else if code == CODE_WRAPPER_NOT_INSTALLED.Code {
		return CODE_WRAPPER_NOT_INSTALLED.HttpStatus
	} else if code == CODE_WRAPPER_SERVER.Code {
		return CODE_WRAPPER_SERVER.HttpStatus
	} else {
		return CODE_WRAPPER_UNKNOWN.HttpStatus
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

	wr := &WebResult{
		Code: codeWrapper.Code,
		Msg:  description,
	}
	return wr
}

//所有的数据库错误情况
var (
	DB_ERROR_DUPLICATE_KEY  = "Error 1062: Duplicate entry"
	DB_ERROR_NOT_FOUND      = "record not found"
	DB_TOO_MANY_CONNECTIONS = "Error 1040: Too many connections"
	DB_BAD_CONNECTION = "driver: bad connection"
)
