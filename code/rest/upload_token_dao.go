package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/nu7hatch/gouuid"
	"time"
)

type UploadTokenDao struct {
	BaseDao
}

//find by uuid. if not found return nil.
func (this *UploadTokenDao) FindByUuid(uuid string) *UploadToken {
	var entity = &UploadToken{}
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
func (this *UploadTokenDao) CheckByUuid(uuid string) *UploadToken {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *UploadTokenDao) Create(uploadToken *UploadToken) *UploadToken {

	timeUUID, _ := uuid.NewV4()
	uploadToken.Uuid = string(timeUUID.String())

	uploadToken.CreateTime = time.Now()
	uploadToken.UpdateTime = time.Now()
	uploadToken.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(uploadToken)
	this.PanicError(db.Error)

	return uploadToken
}

func (this *UploadTokenDao) Save(uploadToken *UploadToken) *UploadToken {

	uploadToken.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(uploadToken)
	this.PanicError(db.Error)

	return uploadToken
}
