package rest

import (
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/robfig/cron"
	"time"
)

//@Service
type DashboardService struct {
	Bean
	dashboardDao  *DashboardDao
	footprintDao  *FootprintDao
	matterDao     *MatterDao
	imageCacheDao *ImageCacheDao
	userDao       *UserDao
}

//初始化方法
func (this *DashboardService) Init() {
	this.Bean.Init()

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := CONTEXT.GetBean(this.dashboardDao)
	if b, ok := b.(*DashboardDao); ok {
		this.dashboardDao = b
	}

	b = CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	b = CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}

}

//系统启动，数据库配置完毕后会调用该方法
func (this *DashboardService) Bootstrap() {

	//每日00:05分清洗离线数据
	expression := "0 5 0 * * ?"
	cronJob := cron.New()
	err := cronJob.AddFunc(expression, this.etl)
	util.PanicError(err)
	cronJob.Start()
	this.logger.Info("[cron job] 每日00:05清洗离线数据")

	//立即执行一次数据清洗任务
	go util.SafeMethod(this.etl)

}

//每日清洗离线数据表。
func (this *DashboardService) etl() {

	this.logger.Info("每日定时数据清洗")

	//准备日期开始结尾
	startTime := util.FirstSecondOfDay(util.Yesterday())
	endTime := util.LastSecondOfDay(util.Yesterday())
	dt := util.ConvertTimeToDateString(startTime)
	longTimeAgo := time.Now()
	longTimeAgo = longTimeAgo.AddDate(-20, 0, 0)

	this.logger.Info("统计汇总表 %s -> %s", util.ConvertTimeToDateTimeString(startTime), util.ConvertTimeToDateTimeString(endTime))

	//判断昨天的记录是否已经生成，如果生成了就直接删除掉
	dbDashboard := this.dashboardDao.FindByDt(dt)
	if dbDashboard != nil {
		this.logger.Info(" %s 的汇总已经存在了，删除以进行更新", dt)
		this.dashboardDao.Delete(dbDashboard)
	}

	invokeNum := this.footprintDao.CountBetweenTime(startTime, endTime)
	this.logger.Info("调用数：%d", invokeNum)

	totalInvokeNum := this.footprintDao.CountBetweenTime(longTimeAgo, endTime)
	this.logger.Info("历史总调用数：%d", totalInvokeNum)

	uv := this.footprintDao.UvBetweenTime(startTime, endTime)
	this.logger.Info("UV：%d", uv)

	totalUv := this.footprintDao.UvBetweenTime(longTimeAgo, endTime)
	this.logger.Info("历史总UV：%d", totalUv)

	matterNum := this.matterDao.CountBetweenTime(startTime, endTime)
	this.logger.Info("文件数量数：%d", matterNum)

	totalMatterNum := this.matterDao.CountBetweenTime(longTimeAgo, endTime)
	this.logger.Info("历史文件总数：%d", totalMatterNum)

	var matterSize int64
	if matterNum != 0 {
		matterSize = this.matterDao.SizeBetweenTime(startTime, endTime)
	}
	this.logger.Info("文件大小：%d", matterSize)

	var totalMatterSize int64
	if totalMatterNum != 0 {
		totalMatterSize = this.matterDao.SizeBetweenTime(longTimeAgo, endTime)
	}
	this.logger.Info("历史文件总大小：%d", totalMatterSize)

	cacheSize := this.imageCacheDao.SizeBetweenTime(startTime, endTime)
	this.logger.Info("缓存大小：%d", cacheSize)

	totalCacheSize := this.imageCacheDao.SizeBetweenTime(longTimeAgo, endTime)
	this.logger.Info("历史缓存总大小：%d", totalCacheSize)

	avgCost := this.footprintDao.AvgCostBetweenTime(startTime, endTime)
	this.logger.Info("平均耗时：%d ms", avgCost)

	dashboard := &Dashboard{
		InvokeNum:      invokeNum,
		TotalInvokeNum: totalInvokeNum,
		Uv:             uv,
		TotalUv:        totalUv,
		MatterNum:      matterNum,
		TotalMatterNum: totalMatterNum,
		FileSize:       matterSize + cacheSize,
		TotalFileSize:  totalMatterSize + totalCacheSize,
		AvgCost:        avgCost,
		Dt:             dt,
	}

	this.dashboardDao.Create(dashboard)

}
