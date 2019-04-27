package rest

import (
	"github.com/eyebluecn/tank/code/core"
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
	db := core.CONTEXT.GetDB().Where(&DownloadToken{Base: Base{Uuid: uuid}}).First(downloadToken)
	if db.Error != nil {
		return nil
	}
	return downloadToken

}

//按照Id查询
func (this *DownloadTokenDao) CheckByUuid(uuid string) *DownloadToken {

	// Read
	var downloadToken = &DownloadToken{}
	db := core.CONTEXT.GetDB().Where(&DownloadToken{Base: Base{Uuid: uuid}}).First(downloadToken)
	this.PanicError(db.Error)
	return downloadToken

}

//创建一个session并且持久化到数据库中。
func (this *DownloadTokenDao) Create(downloadToken *DownloadToken) *DownloadToken {

	timeUUID, _ := uuid.NewV4()
	downloadToken.Uuid = string(timeUUID.String())

	downloadToken.CreateTime = time.Now()
	downloadToken.UpdateTime = time.Now()
	downloadToken.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}

//修改一个downloadToken
func (this *DownloadTokenDao) Save(downloadToken *DownloadToken) *DownloadToken {

	downloadToken.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}

//执行清理操作
func (this *DownloadTokenDao) Cleanup() {
	this.logger.Info("[DownloadTokenDao]执行清理：清除数据库中所有DownloadToken记录。")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(DownloadToken{})
	this.PanicError(db.Error)
}
