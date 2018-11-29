package rest

//@Service
type SecurityVisitService struct {
	Bean
	securityVisitDao *SecurityVisitDao
	userDao          *UserDao
}

//初始化方法
func (this *SecurityVisitService) Init() {

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.securityVisitDao)
	if b, ok := b.(*SecurityVisitDao); ok {
		this.securityVisitDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *SecurityVisitService) Detail(uuid string) *SecurityVisit {

	securityVisit := this.securityVisitDao.CheckByUuid(uuid)

	return securityVisit
}
