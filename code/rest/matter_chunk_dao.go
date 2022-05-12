package rest

type MatterChunkDao struct {
	BaseDao
}

func (this *MatterChunkDao) Init() {
	this.BaseDao.Init()
}

// 通过md5值来查找记录
func (this *MatterChunkDao) FindByMd5(uuid string) *MatterChunk {

}
