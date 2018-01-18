package rest

type WebResult struct {
	Code int       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (this *WebResult) Error() string {
	return this.Msg
}

const (
	//正常
	RESULT_CODE_OK = 200

	//未登录
	RESULT_CODE_LOGIN = -400

	//没有权限
	RESULT_CODE_UNAUTHORIZED = -401

	//请求错误
	RESULT_CODE_BAD_REQUEST = -402

	//没有找到
	RESULT_CODE_NOT_FOUND = -404

	//登录过期
	RESULT_CODE_LOGIN_EXPIRED = -405

	//该登录用户不是有效用户
	RESULT_CODE_LOGIN_INVALID = -406

	//提交的表单验证不通过
	RESULT_CODE_FORM_INVALID = -410

	//请求太频繁
	RESULT_CODE_FREQUENCY = -420

	//服务器出错。
	RESULT_CODE_SERVER_ERROR = -500

	//远程服务不可用
	RESULT_CODE_NOT_AVAILABLE = -501

	//并发异常
	RESULT_CODE_CONCURRENCY = -511

	//远程微服务没有找到
	RESULT_CODE_SERVICE_NOT_FOUND = -600

	//远程微服务连接超时
	RESULT_CODE_SERVICE_TIME_OUT = -610

	//通用的异常
	RESULT_CODE_UTIL_EXCEPTION = -700
)

func ConstWebResult(code int) *WebResult {

	wr := &WebResult{}
	switch code {
	//正常
	case RESULT_CODE_OK:
		wr.Msg = "成功"
		//未登录
	case RESULT_CODE_LOGIN:
		wr.Msg = "没有登录，禁止访问"
		//没有权限
	case RESULT_CODE_UNAUTHORIZED:
		wr.Msg = "没有权限"
		//请求错误
	case RESULT_CODE_BAD_REQUEST:
		wr.Msg = "请求错误"
		//没有找到
	case RESULT_CODE_NOT_FOUND:
		wr.Msg = "没有找到"
		//登录过期
	case RESULT_CODE_LOGIN_EXPIRED:
		wr.Msg = "登录过期"

		//该登录用户不是有效用户
	case RESULT_CODE_LOGIN_INVALID:
		wr.Msg = "该登录用户不是有效用户或者用户已被禁用"

		//提交的表单验证不通过
	case RESULT_CODE_FORM_INVALID:
		wr.Msg = "提交的表单验证不通过"
		//请求太频繁
	case RESULT_CODE_FREQUENCY:
		wr.Msg = "请求太频繁"
		//服务器出错。
	case RESULT_CODE_SERVER_ERROR:
		wr.Msg = "服务器出错"
		//远程服务不可用
	case RESULT_CODE_NOT_AVAILABLE:
		wr.Msg = "远程服务不可用"
		//并发异常
	case RESULT_CODE_CONCURRENCY:
		wr.Msg = "并发异常"
		//远程微服务没有找到
	case RESULT_CODE_SERVICE_NOT_FOUND:
		wr.Msg = "远程微服务没有找到"
		//远程微服务连接超时
	case RESULT_CODE_SERVICE_TIME_OUT:
		wr.Msg = "远程微服务连接超时"
	default:
		code = RESULT_CODE_UTIL_EXCEPTION
		wr.Msg = "服务器未知错误"
	}
	wr.Code = code
	return wr

}
