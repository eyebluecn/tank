package code

//@Service
type PreferenceService struct {
	Bean
	preferenceDao *PreferenceDao
	preference    *Preference
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

//获取单例的配置。
func (this *PreferenceService) Fetch() *Preference {

	if this.preference == nil {
		this.preference = this.preferenceDao.Fetch()
	}

	return this.preference
}

//清空单例配置。
func (this *PreferenceService) Reset() {

	this.preference = nil

}

//执行清理操作
func (this *PreferenceService) Cleanup() {

	this.logger.Info("[PreferenceService]执行清理：重置缓存中的preference。")

	this.Reset()
}
