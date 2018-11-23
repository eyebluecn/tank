package rest

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nu7hatch/gouuid"
	"time"
)

type DownloadTokenDao struct {
	BaseDao
}

//按照Id查询
func (this *DownloadTokenDao) FindByUuid(uuid string) *DownloadToken {

	// Read
	var downloadToken = &DownloadToken{}
	db := this.context.DB.Where(&DownloadToken{Base: Base{Uuid: uuid}}).First(downloadToken)
	if db.Error != nil {
		return nil
	}
	return downloadToken

}

//按照Id查询
func (this *DownloadTokenDao) CheckByUuid(uuid string) *DownloadToken {

	// Read
	var downloadToken = &DownloadToken{}
	db := this.context.DB.Where(&DownloadToken{Base: Base{Uuid: uuid}}).First(downloadToken)
	this.PanicError(db.Error)
	return downloadToken

}

//创建一个session并且持久化到数据库中。
func (this *DownloadTokenDao) Create(downloadToken *DownloadToken) *DownloadToken {

	timeUUID, _ := uuid.NewV4()
	downloadToken.Uuid = string(timeUUID.String())

	downloadToken.CreateTime = time.Now()
	downloadToken.UpdateTime = time.Now()

	db := this.context.DB.Create(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}

//修改一个downloadToken
func (this *DownloadTokenDao) Save(downloadToken *DownloadToken) *DownloadToken {

	downloadToken.UpdateTime = time.Now()
	db := this.context.DB.Save(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}
