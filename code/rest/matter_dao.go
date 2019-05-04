package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/jinzhu/gorm"
	"github.com/nu7hatch/gouuid"
	"os"
	"time"
)

type MatterDao struct {
	BaseDao
	imageCacheDao *ImageCacheDao
	bridgeDao     *BridgeDao
}

//初始化方法
func (this *MatterDao) Init() {
	this.BaseDao.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

}

//按照Id查询文件
func (this *MatterDao) FindByUuid(uuid string) *Matter {

	// Read
	var matter Matter
	db := core.CONTEXT.GetDB().Where(&Matter{Base: Base{Uuid: uuid}}).First(&matter)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}
	return &matter
}

//按照Id查询文件
func (this *MatterDao) CheckByUuid(uuid string) *Matter {
	matter := this.FindByUuid(uuid)
	if matter == nil {
		panic(result.NotFound("%s 对应的matter不存在", uuid))
	}
	return matter
}

//按照uuid查找一个文件夹，可能返回root对应的matter.
func (this *MatterDao) CheckWithRootByUuid(uuid string, user *User) *Matter {

	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null."))
	}

	var matter *Matter
	if uuid == MATTER_ROOT {
		if user == nil {
			panic(result.BadRequest("user cannot be null."))
		}
		matter = NewRootMatter(user)
	} else {
		matter = this.CheckByUuid(uuid)
	}

	return matter
}

//按照path查找一个matter，可能返回root对应的matter.
func (this *MatterDao) CheckWithRootByPath(path string, user *User) *Matter {

	var matter *Matter

	if user == nil {
		panic(result.BadRequest("user cannot be null."))
	}

	//目标文件夹matter
	if path == "" || path == "/" {
		matter = NewRootMatter(user)
	} else {
		matter = this.checkByUserUuidAndPath(user.Uuid, path)
	}

	return matter
}

//按照名字查询文件夹
func (this *MatterDao) FindByUserUuidAndPuuidAndNameAndDirTrue(userUuid string, puuid string, name string) *Matter {

	var wp = &builder.WherePair{}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{1}})

	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		return nil
	}

	return matter
}

//按照id和userUuid来查找。找不到抛异常。
func (this *MatterDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *Matter {

	// Read
	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Where(&Matter{Base: Base{Uuid: uuid}, UserUuid: userUuid}).First(matter)
	this.PanicError(db.Error)

	return matter

}

//统计某个用户的某个文件夹下的某个名字的文件(或文件夹)数量。
func (this *MatterDao) CountByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) int {

	var matter Matter
	var count int

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{dir}})

	db := core.CONTEXT.GetDB().
		Model(&matter).
		Where(wp.Query, wp.Args...).
		Count(&count)
	this.PanicError(db.Error)

	return count
}

//统计某个用户的某个文件夹下的某个名字的文件(或文件夹)数量。
func (this *MatterDao) FindByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) *Matter {

	var matter = &Matter{}

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []interface{}{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{dir}})

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

//获取某个用户的某个文件夹下的某个名字的文件(或文件夹)列表
func (this *MatterDao) ListByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) []*Matter {

	var matters []*Matter

	db := core.CONTEXT.GetDB().
		Where(Matter{UserUuid: userUuid, Puuid: puuid, Dir: dir, Name: name}).
		Find(&matters)
	this.PanicError(db.Error)

	return matters
}

//获取某个文件夹下所有的文件和子文件
func (this *MatterDao) ListByPuuidAndUserUuid(puuid string, userUuid string, sortArray []builder.OrderPair) []*Matter {
	var matters []*Matter

	if sortArray == nil {

		//顺序按照文件夹，创建时间
		sortArray = []builder.OrderPair{
			{
				Key:   "dir",
				Value: DIRECTION_DESC,
			},
			{
				Key:   "create_time",
				Value: DIRECTION_DESC,
			},
		}
	}

	db := core.CONTEXT.GetDB().Where(Matter{UserUuid: userUuid, Puuid: puuid}).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}

//根据uuid查找对应的Matters
func (this *MatterDao) ListByUuids(uuids []string, sortArray []builder.OrderPair) []*Matter {
	var matters []*Matter

	db := core.CONTEXT.GetDB().Where(uuids).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}

//获取某个文件夹下所有的文件和子文件
func (this *MatterDao) Page(page int, pageSize int, puuid string, userUuid string, name string, dir string, extensions []string, sortArray []builder.OrderPair) *Pager {

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []interface{}{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []interface{}{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name LIKE ?", Args: []interface{}{"%" + name + "%"}})
	}

	if dir == TRUE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{1}})
	} else if dir == FALSE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []interface{}{0}})
	}

	var conditionDB *gorm.DB
	if extensions != nil && len(extensions) > 0 {
		var orWp = &builder.WherePair{}

		for _, v := range extensions {
			orWp = orWp.Or(&builder.WherePair{Query: "name LIKE ?", Args: []interface{}{"%." + v}})
		}

		conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Where(orWp.Query, orWp.Args...)
	} else {
		conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...)
	}

	count := 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var matters []*Matter
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&matters)
	this.PanicError(db.Error)
	pager := NewPager(page, pageSize, count, matters)

	return pager
}

//创建
func (this *MatterDao) Create(matter *Matter) *Matter {

	timeUUID, _ := uuid.NewV4()
	matter.Uuid = string(timeUUID.String())
	matter.CreateTime = time.Now()
	matter.UpdateTime = time.Now()
	matter.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(matter)
	this.PanicError(db.Error)

	return matter
}

//修改一个文件
func (this *MatterDao) Save(matter *Matter) *Matter {

	matter.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(matter)
	this.PanicError(db.Error)

	return matter
}

//计数器加一
func (this *MatterDao) TimesIncrement(matterUuid string) {
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matterUuid).Update("times", gorm.Expr("times + 1"))
	this.PanicError(db.Error)
}

//获取一个文件夹中直系文件/文件夹的总大小 puuid可以传root
func (this *MatterDao) SizeByPuuidAndUserUuid(matterUuid string, userUuid string) int64 {

	var wp = &builder.WherePair{Query: "puuid = ? AND user_uuid = ?", Args: []interface{}{matterUuid, userUuid}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize
}

//统计某个文件/文件夹的大小(会自动往上统计，直到根目录)
func (this *MatterDao) ComputeRouteSize(matterUuid string, userUuid string) {

	//如果更新到了根目录，那么更新到用户身上。
	if matterUuid == MATTER_ROOT {

		size := this.SizeByPuuidAndUserUuid(MATTER_ROOT, userUuid)

		//更新用户文件的总大小。
		db := core.CONTEXT.GetDB().Model(&User{}).Where("uuid = ?", userUuid).Update("total_size", size)
		this.PanicError(db.Error)

		return
	}

	matter := this.CheckByUuid(matterUuid)

	//只有文件夹才去统计
	if matter.Dir {
		//计算该目录下的直系文件/文件夹总大小
		size := this.SizeByPuuidAndUserUuid(matterUuid, userUuid)

		//大小有变化才更新
		if matter.Size != size {
			//更新大小。
			db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matterUuid).Update("size", size)
			this.PanicError(db.Error)
		}

	}

	//更新自己的上一级目录。
	this.ComputeRouteSize(matter.Puuid, userUuid)
}

//删除一个文件，数据库中删除，物理磁盘上删除。
func (this *MatterDao) Delete(matter *Matter) {

	//目录的话递归删除。
	if matter.Dir {
		matters := this.ListByPuuidAndUserUuid(matter.Uuid, matter.UserUuid, nil)

		for _, f := range matters {
			this.Delete(f)
		}

		//删除数据库中文件夹本身
		db := core.CONTEXT.GetDB().Delete(&matter)
		this.PanicError(db.Error)

		//从磁盘中删除该文件夹。
		util.DeleteEmptyDir(matter.AbsolutePath())

	} else {

		//删除数据库中文件记录
		db := core.CONTEXT.GetDB().Delete(&matter)
		this.PanicError(db.Error)

		//删除对应的缓存图片。
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		//删除所有的分享文件
		this.bridgeDao.DeleteByMatterUuid(matter.Uuid)

		//删除文件
		err := os.Remove(matter.AbsolutePath())
		if err != nil {
			this.logger.Error("删除磁盘上的文件出错 %s", err.Error())
		}

		//由于目录和物理结构一一对应，这里不能删除上级文件夹。

	}
}

//获取一段时间中，总的数量
func (this *MatterDao) CountBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Count(&count)
	this.PanicError(db.Error)
	return count
}

//获取一段时间中文件总大小
func (this *MatterDao) SizeBetweenTime(startTime time.Time, endTime time.Time) int64 {

	//TODO: 所有函数汇总的SQL均需要先count询问，再处理。
	var size int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&size)
	this.PanicError(err)
	return size
}

//根据userUuid和path来查找
func (this *MatterDao) findByUserUuidAndPath(userUuid string, path string) *Matter {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path = ?", Args: []interface{}{userUuid, path}}

	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

//根据userUuid和path来查找
func (this *MatterDao) checkByUserUuidAndPath(userUuid string, path string) *Matter {

	if path == "" {
		panic(result.BadRequest("path cannot be null"))
	}
	matter := this.findByUserUuidAndPath(userUuid, path)
	if matter == nil {
		panic(result.NotFound("path = %s not exists", path))
	}

	return matter
}

//获取一个文件夹中文件总大小
func (this *MatterDao) SumSizeByUserUuidAndPath(userUuid string, path string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path like ?", Args: []interface{}{userUuid, path + "%"}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize

}

//一个文件夹中的数量
func (this *MatterDao) CountByUserUuidAndPath(userUuid string, path string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path like ?", Args: []interface{}{userUuid, path + "%"}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	core.PanicError(db.Error)

	return count

}

//执行清理操作
func (this *MatterDao) Cleanup() {
	this.logger.Info("[MatterDao]执行清理：清除数据库中所有Matter记录。删除磁盘中所有Matter文件。")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Matter{})
	this.PanicError(db.Error)

	err := os.RemoveAll(core.CONFIG.MatterPath())
	this.PanicError(err)

}
