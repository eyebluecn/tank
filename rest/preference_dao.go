package rest

import (
	"github.com/nu7hatch/gouuid"
	"time"
)

type PreferenceDao struct {
	BaseDao
}

//按照Id查询偏好设置
func (this *PreferenceDao) Fetch() *Preference {

	// Read
	var preference = &Preference{}
	db := CONTEXT.DB.First(preference)
	if db.Error != nil {

		if db.Error.Error() == "record not found" {
			preference.Name = "蓝眼云盘"
			preference.ShowAlien = true
			this.Create(preference)
			return preference
		} else {
			return nil
		}

	}

	return preference
}

//创建
func (this *PreferenceDao) Create(preference *Preference) *Preference {

	timeUUID, _ := uuid.NewV4()
	preference.Uuid = string(timeUUID.String())
	preference.CreateTime = time.Now()
	preference.UpdateTime = time.Now()
	preference.Sort = time.Now().UnixNano() / 1e6
	db := CONTEXT.DB.Create(preference)
	this.PanicError(db.Error)

	return preference
}

//修改一个偏好设置
func (this *PreferenceDao) Save(preference *Preference) *Preference {

	preference.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(preference)
	this.PanicError(db.Error)

	return preference
}
