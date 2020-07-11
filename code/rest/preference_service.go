package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

//@Service
type PreferenceService struct {
	BaseBean
	preferenceDao *PreferenceDao
	preference    *Preference
	matterDao     *MatterDao
	matterService *MatterService
	userDao       *UserDao
	migrating     bool
}

func (this *PreferenceService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.preferenceDao)
	if b, ok := b.(*PreferenceDao); ok {
		this.preferenceDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

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

//清空单例配置。
func (this *PreferenceService) Save(preference *Preference) *Preference {

	preference = this.preferenceDao.Save(preference)

	//clean cache.
	this.Reset()

	return preference
}

//System cleanup.
func (this *PreferenceService) Cleanup() {

	this.logger.Info("[PreferenceService] clean up. Delete all preference")

	this.Reset()
}
