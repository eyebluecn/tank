package rest

import (

	"github.com/nu7hatch/gouuid"
	"time"
)

type UserDao struct {
	BaseDao
}

//创建用户
func (this *UserDao) Create(user *User) *User {

	if user == nil {
		panic("参数不能为nil")
	}

	timeUUID, _ := uuid.NewV4()
	user.Uuid = string(timeUUID.String())
	user.CreateTime = time.Now()
	user.UpdateTime = time.Now()
	user.LastTime = time.Now()
	user.Sort = time.Now().UnixNano() / 1e6

	db := CONTEXT.DB.Create(user)
	this.PanicError(db.Error)

	return user
}

//按照Id查询用户，找不到返回nil
func (this *UserDao) FindByUuid(uuid string) *User {

	// Read
	var user *User = &User{}
	db := CONTEXT.DB.Where(&User{Base: Base{Uuid: uuid}}).First(user)
	if db.Error != nil {
		return nil
	}
	return user
}

//按照Id查询用户,找不到抛panic
func (this *UserDao) CheckByUuid(uuid string) *User {

	if uuid == "" {
		panic("uuid必须指定")
	}

	// Read
	var user *User = &User{}
	db := CONTEXT.DB.Where(&User{Base: Base{Uuid: uuid}}).First(user)
	this.PanicError(db.Error)
	return user
}

//按照邮箱查询用户。
func (this *UserDao) FindByEmail(email string) *User {

	var user *User = &User{}
	db := CONTEXT.DB.Where(&User{Email: email}).First(user)
	if db.Error != nil {
		return nil
	}
	return user
}

//显示用户列表。
func (this *UserDao) Page(page int, pageSize int, username string, email string, phone string, status string, sortArray []OrderPair) *Pager {

	var wp = &WherePair{}

	if username != "" {
		wp = wp.And(&WherePair{Query: "username LIKE ?", Args: []interface{}{"%" + username + "%"}})
	}

	if email != "" {
		wp = wp.And(&WherePair{Query: "email LIKE ?", Args: []interface{}{"%" + email + "%"}})
	}

	if phone != "" {
		wp = wp.And(&WherePair{Query: "phone = ?", Args: []interface{}{phone}})
	}

	if status != "" {
		wp = wp.And(&WherePair{Query: "status = ?", Args: []interface{}{status}})
	}

	count := 0
	db := CONTEXT.DB.Model(&User{}).Where(wp.Query, wp.Args...).Count(&count)
	this.PanicError(db.Error)

	var users []*User
	orderStr := this.GetSortString(sortArray)
	if orderStr == "" {
		db = CONTEXT.DB.Where(wp.Query, wp.Args...).Offset(page * pageSize).Limit(pageSize).Find(&users)
	} else {
		db = CONTEXT.DB.Where(wp.Query, wp.Args...).Order(orderStr).Offset(page * pageSize).Limit(pageSize).Find(&users)
	}

	this.PanicError(db.Error)

	pager := NewPager(page, pageSize, count, users)

	return pager
}

//查询某个用户名是否已经有用户了
func (this *UserDao) CountByUsername(username string) int {
	var count int
	db := CONTEXT.DB.
		Model(&User{}).
		Where("username = ?", username).
		Count(&count)
	this.PanicError(db.Error)
	return count
}

//查询某个邮箱是否已经有用户了
func (this *UserDao) CountByEmail(email string) int {
	var count int
	db := CONTEXT.DB.
		Model(&User{}).
		Where("email = ?", email).
		Count(&count)
	this.PanicError(db.Error)
	return count
}

//保存用户
func (this *UserDao) Save(user *User) *User {

	user.UpdateTime = time.Now()
	db := CONTEXT.DB.
		Save(user)
	this.PanicError(db.Error)
	return user
}
