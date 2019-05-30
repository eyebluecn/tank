package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/nu7hatch/gouuid"
	"time"
)

type UserDao struct {
	BaseDao
}

func (this *UserDao) Init() {
	this.BaseDao.Init()
}

func (this *UserDao) Create(user *User) *User {

	if user == nil {
		panic(result.BadRequest("user cannot be nil"))
	}

	timeUUID, _ := uuid.NewV4()
	user.Uuid = string(timeUUID.String())
	user.CreateTime = time.Now()
	user.UpdateTime = time.Now()
	user.LastTime = time.Now()
	user.Sort = time.Now().UnixNano() / 1e6

	db := core.CONTEXT.GetDB().Create(user)
	this.PanicError(db.Error)

	return user
}

//find by uuid. if not found return nil.
func (this *UserDao) FindByUuid(uuid string) *User {
	var entity = &User{}
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
func (this *UserDao) CheckByUuid(uuid string) *User {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *UserDao) FindByUsername(username string) *User {

	var user = &User{}
	db := core.CONTEXT.GetDB().Where(&User{Username: username}).First(user)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return user
}

func (this *UserDao) Page(page int, pageSize int, username string, status string, sortArray []builder.OrderPair) *Pager {

	var wp = &builder.WherePair{}

	if username != "" {
		wp = wp.And(&builder.WherePair{Query: "username LIKE ?", Args: []interface{}{"%" + username + "%"}})
	}

	if status != "" {
		wp = wp.And(&builder.WherePair{Query: "status = ?", Args: []interface{}{status}})
	}

	count := 0
	db := core.CONTEXT.GetDB().Model(&User{}).Where(wp.Query, wp.Args...).Count(&count)
	this.PanicError(db.Error)

	var users []*User
	orderStr := this.GetSortString(sortArray)
	if orderStr == "" {
		db = core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).Offset(page * pageSize).Limit(pageSize).Find(&users)
	} else {
		db = core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).Order(orderStr).Offset(page * pageSize).Limit(pageSize).Find(&users)
	}

	this.PanicError(db.Error)

	pager := NewPager(page, pageSize, count, users)

	return pager
}

func (this *UserDao) CountByUsername(username string) int {
	var count int
	db := core.CONTEXT.GetDB().
		Model(&User{}).
		Where("username = ?", username).
		Count(&count)
	this.PanicError(db.Error)
	return count
}

func (this *UserDao) Save(user *User) *User {

	user.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().
		Save(user)
	this.PanicError(db.Error)
	return user
}

//find all 2.0 users.
func (this *UserDao) FindUsers20() []*User {
	var users []*User
	var wp = &builder.WherePair{}
	wp = wp.And(&builder.WherePair{Query: "username like ?", Args: []interface{}{"%_20"}})

	db := core.CONTEXT.GetDB().Model(&User{}).Where(wp.Query, wp.Args...).Find(&users)
	this.PanicError(db.Error)
	return users
}

func (this *UserDao) DeleteUsers20() {
	var wp = &builder.WherePair{}
	wp = wp.And(&builder.WherePair{Query: "username like ?", Args: []interface{}{"%_20"}})

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).Delete(User{})
	this.PanicError(db.Error)
}

//System cleanup.
func (this *UserDao) Cleanup() {
	this.logger.Info("[UserDao] clean up. Delete all User")
	db := core.CONTEXT.GetDB().Where("uuid is not null and role != ?", USER_ROLE_ADMINISTRATOR).Delete(User{})
	this.PanicError(db.Error)
}
