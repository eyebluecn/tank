package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/nu7hatch/gouuid"
	"time"
)

type DownloadTokenDao struct {
	BaseDao
}

//find by uuid. if not found return nil.
func (this *DownloadTokenDao) FindByUuid(uuid string) *DownloadToken {
	var entity = &DownloadToken{}
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
func (this *DownloadTokenDao) CheckByUuid(uuid string) *DownloadToken {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

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

func (this *DownloadTokenDao) Save(downloadToken *DownloadToken) *DownloadToken {

	downloadToken.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(downloadToken)
	this.PanicError(db.Error)

	return downloadToken
}

func (this *DownloadTokenDao) Cleanup() {
	this.logger.Info("[DownloadTokenDao] clean up. Delete all DownloadToken")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(DownloadToken{})
	this.PanicError(db.Error)
}
