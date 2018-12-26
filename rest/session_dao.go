package rest

import (

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

//创建一个session并且持久化到数据库中。
func (this *SessionDao) Create(session *Session) *Session {

	timeUUID, _ := uuid.NewV4()
	session.Uuid = string(timeUUID.String())
	session.CreateTime = time.Now()
	session.UpdateTime = time.Now()
	session.Sort = time.Now().UnixNano() / 1e6
	db := CONTEXT.DB.Create(session)
	this.PanicError(db.Error)

	return session
}


//修改一个session
func (this *SessionDao) Save(session *Session) *Session {

	session.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(session)
	this.PanicError(db.Error)

	return session
}


func (this *SessionDao) Delete(uuid string) {

	session := this.CheckByUuid(uuid)

	session.ExpireTime = time.Now()
	db := CONTEXT.DB.Delete(session)

	this.PanicError(db.Error)

}


