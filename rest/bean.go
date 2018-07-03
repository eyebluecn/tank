package rest

import "net/http"

type IBean interface {
	Init(context *Context)
	PanicError(err error);
	PanicWebError(msg string, code int);
}

type Bean struct {
	context *Context
}

func (this *Bean) Init(context *Context) {
	this.context = context
}

//处理错误的统一方法
func (this *Bean) PanicError(err error) {
	if err != nil {
		panic(&WebError{Msg: err.Error(), Code: http.StatusInternalServerError})
	}
}

//处理错误的统一方法
func (this *Bean) PanicWebError(msg string, httpStatusCode int) {
	panic(&WebError{Msg: msg, Code: httpStatusCode})
}
