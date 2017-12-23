package rest

type IBean interface {
	Init(context *Context)
	PanicError(err error);
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
		panic(err)
	}
}
