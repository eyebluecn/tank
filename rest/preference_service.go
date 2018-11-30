package rest

//@Service
type PreferenceService struct {
	Bean
	preferenceDao *PreferenceDao
}

//初始化方法
func (this *PreferenceService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*PreferenceDao); ok {
		this.preferenceDao = b
	}

}
