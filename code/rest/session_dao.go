package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/nu7hatch/gouuid"
	"time"
)

type SessionDao struct {
	BaseDao
}

//find by uuid. if not found return nil.
func (this *SessionDao) FindByUuid(uuid string) *Session {
	var entity = &Session{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", uuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

//find by uuid. if not found panic NotFound error
func (this *SessionDao) CheckByUuid(uuid string) *Session {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *SessionDao) Create(session *Session) *Session {

	timeUUID, _ := uuid.NewV4()
	session.Uuid = string(timeUUID.String())
	session.CreateTime = time.Now()
	session.UpdateTime = time.Now()
	session.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(session)
	this.PanicError(db.Error)

	return session
}

func (this *SessionDao) Save(session *Session) *Session {

	session.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(session)
	this.PanicError(db.Error)

	return session
}

func (this *SessionDao) Delete(uuid string) {

	session := this.CheckByUuid(uuid)

	session.ExpireTime = time.Now()
	db := core.CONTEXT.GetDB().Delete(session)

	this.PanicError(db.Error)

}

//System cleanup.
func (this *SessionDao) Cleanup() {
	this.logger.Info("[SessionDao] clean up. Delete all Session")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Session{})
	this.PanicError(db.Error)
}
