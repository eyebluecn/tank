package rest

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
	db := this.context.DB.Where(&UploadToken{Base: Base{Uuid: uuid}}).First(uploadToken)
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
	uploadToken.ModifyTime = time.Now()

	db := this.context.DB.Create(uploadToken)
	this.PanicError(db.Error)

	return uploadToken
}

//修改一个uploadToken
func (this *UploadTokenDao) Save(uploadToken *UploadToken) *UploadToken {

	uploadToken.ModifyTime = time.Now()
	db := this.context.DB.Save(uploadToken)
	this.PanicError(db.Error)

	return uploadToken
}
