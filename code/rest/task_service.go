package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/robfig/cron/v3"
	"net/http"
)

// system tasks service
// @Service
type TaskService struct {
	BaseBean
	footprintService  *FootprintService
	dashboardService  *DashboardService
	preferenceService *PreferenceService
	matterService     *MatterService
	userDao           *UserDao
	spaceDao          *SpaceDao

	//whether scan task is running
	scanTaskRunning bool
	scanTaskCron    *cron.Cron
}

func (this *TaskService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.footprintService)
	if b, ok := b.(*FootprintService); ok {
		this.footprintService = b
	}

	b = core.CONTEXT.GetBean(this.dashboardService)
	if b, ok := b.(*DashboardService); ok {
		this.dashboardService = b
	}

	b = core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}
	b = core.CONTEXT.GetBean(this.userDao)
	if b, ok := b.(*UserDao); ok {
		this.userDao = b
	}
	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

	this.scanTaskRunning = false
}

// init the clean footprint task.
func (this *TaskService) InitCleanFootprintTask() {

	//use standard cron expression. 5 fields. ()
	expression := "10 0 * * *"
	cronJob := cron.New()
	_, err := cronJob.AddFunc(expression, this.footprintService.CleanOldData)
	core.PanicError(err)
	cronJob.Start()

	this.logger.Info("[cron job] Every day 00:10 delete Footprint data of 8 days ago.")
}

// init the elt task.
func (this *TaskService) InitEtlTask() {

	expression := "5 0 * * *"
	cronJob := cron.New()
	_, err := cronJob.AddFunc(expression, this.dashboardService.Etl)
	core.PanicError(err)
	cronJob.Start()

	this.logger.Info("[cron job] Everyday 00:05 ETL dashboard data.")
}

// init the clean deleted matters task.
func (this *TaskService) InitCleanDeletedMattersTask() {

	expression := "0 1 * * *"
	cronJob := cron.New()
	_, err := cronJob.AddFunc(expression, this.matterService.CleanExpiredDeletedMatters)
	core.PanicError(err)
	cronJob.Start()

	this.logger.Info("[cron job] Everyday 01:00 Clean deleted matters.")
}

// scan task.
func (this *TaskService) doScanTask() {

	if this.scanTaskRunning {
		this.logger.Info("scan task is processing. Give up this invoke.")
		return
	} else {
		this.scanTaskRunning = true
	}

	defer func() {
		if err := recover(); err != nil {
			this.logger.Info("occur error when do scan task.")
		}
		this.logger.Info("finish the scan task.")
		this.scanTaskRunning = false
	}()

	this.logger.Info("[cron job] do the scan task.")
	preference := this.preferenceService.Fetch()
	scanConfig := preference.FetchScanConfig()

	if !scanConfig.Enable {
		this.logger.Info("scan task not enabled.")
		return
	}

	//mock a request.
	request := &http.Request{}

	if scanConfig.Scope == SCAN_SCOPE_ALL {
		//scan all user's root folder.
		this.spaceDao.PageHandle(func(space *Space) {

			core.RunWithRecovery(func() {

				this.logger.Info("scan spaceName = %s", space.Name)

				//find user by space
				user := this.userDao.FindByUuid(space.UserUuid)
				if user == nil {
					user = this.userDao.FindAnAdmin()
				}
				this.matterService.DeleteByPhysics(request, user, space)
				this.matterService.ScanPhysics(request, user, space)

			})

		})

	} else if scanConfig.Scope == SCAN_SCOPE_CUSTOM {

		for _, spaceName := range scanConfig.SpaceNames {
			space := this.spaceDao.FindByName(spaceName)
			if space == nil {
				this.logger.Error("name = %s not exist.", spaceName)
			} else {
				this.logger.Info("scan custom user folder. spaceName = %s", spaceName)

				core.RunWithRecovery(func() {
					//find user by space
					user := this.userDao.FindByUuid(space.UserUuid)
					if user == nil {
						user = this.userDao.FindAnAdmin()
					}
					this.matterService.DeleteByPhysics(request, user, space)
					this.matterService.ScanPhysics(request, user, space)

				})

			}
		}
	}

}

// init the scan task.
func (this *TaskService) InitScanTask() {

	if this.scanTaskCron != nil {
		this.scanTaskCron.Stop()
		this.scanTaskCron = nil
	}

	preference := this.preferenceService.Fetch()
	scanConfig := preference.FetchScanConfig()

	if !scanConfig.Enable {
		this.logger.Info("scan task not enabled.")
		return
	}

	if !util.ValidateCron(scanConfig.Cron) {
		this.logger.Info("cron spec %s error", scanConfig.Cron)
		return
	}

	this.scanTaskCron = cron.New()
	_, err := this.scanTaskCron.AddFunc(scanConfig.Cron, this.doScanTask)
	core.PanicError(err)
	this.scanTaskCron.Start()

	this.logger.Info("[cron job] %s do scan task.", scanConfig.Cron)
}

func (this *TaskService) Bootstrap() {

	//load the clean footprint task.
	this.InitCleanFootprintTask()

	//load the etl task.
	this.InitEtlTask()

	//load the clean deleted matters task.
	this.InitCleanDeletedMattersTask()

	//load the scan task.
	this.InitScanTask()

}
