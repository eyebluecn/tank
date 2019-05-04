package rest

import "github.com/eyebluecn/tank/code/core"

//@Service
type PreferenceService struct {
	BaseBean
	preferenceDao *PreferenceDao
	preference    *Preference
}

func (this *PreferenceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.preferenceDao)
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

//System cleanup.
func (this *PreferenceService) Cleanup() {

	this.logger.Info("[PreferenceService] clean up. Delete all preference")

	this.Reset()
}
