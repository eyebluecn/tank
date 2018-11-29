package rest

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nu7hatch/gouuid"
	"time"
)

type SessionDao struct {
	BaseDao
}

//构造函数
func NewSessionDao() *SessionDao {

	var sessionDao = &SessionDao{}
	sessionDao.Init()
	return sessionDao
}

//按照Id查询session.
func (this *SessionDao) FindByUuid(uuid string) *Session {

	// Read
	var session = &Session{}
	db := CONTEXT.DB.Where(&Session{Base: Base{Uuid: uuid}}).First(session)
	if db.Error != nil {
		return nil
	}
	return session
}

//按照Id查询session.
func (this *SessionDao) CheckByUuid(uuid string) *Session {

	// Read
	var session = &Session{}
	db := CONTEXT.DB.Where(&Session{Base: Base{Uuid: uuid}}).First(session)
	this.PanicError(db.Error)
	return session
}

//按照authentication查询用户。
func (this *SessionDao) FindByAuthentication(authentication string) *Session {

	var session = &Session{}
	db := CONTEXT.DB.Where(&Session{Authentication: authentication}).First(session)
	if db.Error != nil {
		return nil
	}
	return session

}

//创建一个session并且持久化到数据库中。
func (this *SessionDao) Create(session *Session) *Session {

	timeUUID, _ := uuid.NewV4()
	session.Uuid = string(timeUUID.String())
	db := CONTEXT.DB.Create(session)
	this.PanicError(db.Error)

	return session
}

func (this *SessionDao) Delete(uuid string) {

	session := this.CheckByUuid(uuid)

	session.ExpireTime = time.Now()
	db := CONTEXT.DB.Delete(session)

	this.PanicError(db.Error)

}
