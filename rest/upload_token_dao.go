package rest

import (

	"github.com/nu7hatch/gouuid"
	"time"
)

type UploadTokenDao struct {
	BaseDao
}

//按照Id查询
func (this *UploadTokenDao) FindByUuid(uuid string) *UploadToken {

	// Read
	var uploadToken = &UploadToken{}
	db := CONTEXT.DB.Where(&UploadToken{Base: Base{Uuid: uuid}}).First(uploadToken)
	if db.Error != nil {
		return nil
	}

	return uploadToken

}

//创建一个session并且持久化到数据库中。
func (this *UploadTokenDao) Create(uploadToken *UploadToken) *UploadToken {

	timeUUID, _ := uuid.NewV4()
	uploadToken.Uuid = string(timeUUID.String())

	uploadToken.CreateTime = time.Now()
	uploadToken.UpdateTime = time.Now()
	uploadToken.Sort = time.Now().UnixNano() / 1e6
	db := CONTEXT.DB.Create(uploadToken)
	this.PanicError(db.Error)

	return uploadToken
}

//修改一个uploadToken
func (this *UploadTokenDao) Save(uploadToken *UploadToken) *UploadToken {

	uploadToken.UpdateTime = time.Now()
	db := CONTEXT.DB.Save(uploadToken)
	this.PanicError(db.Error)

	return uploadToken
}
