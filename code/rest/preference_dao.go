package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/nu7hatch/gouuid"
	"time"
)

type PreferenceDao struct {
	BaseDao
}

//find by uuid. if not found return nil.
func (this *PreferenceDao) Fetch() *Preference {

	// Read
	var preference = &Preference{}
	db := core.CONTEXT.GetDB().First(preference)
	if db.Error != nil {

		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			preference.Name = "EyeblueTank"
			preference.Version = core.VERSION
			this.Create(preference)
			return preference
		} else {
			return nil
		}
	}

	preference.Version = core.VERSION
	return preference
}

func (this *PreferenceDao) Create(preference *Preference) *Preference {

	timeUUID, _ := uuid.NewV4()
	preference.Uuid = string(timeUUID.String())
	preference.CreateTime = time.Now()
	preference.UpdateTime = time.Now()
	preference.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(preference)
	this.PanicError(db.Error)

	return preference
}

func (this *PreferenceDao) Save(preference *Preference) *Preference {

	preference.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(preference)
	this.PanicError(db.Error)

	return preference
}

//System cleanup.
func (this *PreferenceDao) Cleanup() {

	this.logger.Info("[PreferenceDao] clean up. Delete all Preference")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Preference{})
	this.PanicError(db.Error)
}
