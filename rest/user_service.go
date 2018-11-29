package rest

import (
	"net/http"
)

//@Service
type UserService struct {
	Bean
	userDao *UserDao
}

//初始化方法
func (this *UserService) Init() {

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//装载session信息，如果session没有了根据cookie去装填用户信息。
//在所有的路由最初会调用这个方法
func (this *UserService) enter(writer http.ResponseWriter, request *http.Request) {


}
