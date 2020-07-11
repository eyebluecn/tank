package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/robfig/cron/v3"
)

// system tasks service
//@Service
type TaskService struct {
	BaseBean
	footprintService *FootprintService
	dashboardService *DashboardService
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

}

//init the clean footprint task.
func (this *TaskService) InitCleanFootprintTask() {

	//use standard cron expression. 5 fields. ()
	expression := "10 0 * * *"
	cronJob := cron.New()
	entryId, err := cronJob.AddFunc(expression, this.footprintService.CleanOldData)
	core.PanicError(err)
	cronJob.Start()

	this.logger.Info("[cron job] Every day 00:10 delete Footprint data of 8 days ago. entryId = %d", entryId)
}

//init the elt task.
func (this *TaskService) InitEtlTask() {

	expression := "5 0 * * *"
	cronJob := cron.New()
	entryId, err := cronJob.AddFunc(expression, this.dashboardService.Etl)
	core.PanicError(err)
	cronJob.Start()

	this.logger.Info("[cron job] Everyday 00:05 ETL dashboard data. entryId = %d", entryId)
}

//init the scan task.
func (this *TaskService) InitScanTask() {

	expression := "15 0 * * *"
	cronJob := cron.New()
	entryId, err := cronJob.AddFunc(expression, this.dashboardService.Etl)
	core.PanicError(err)
	cronJob.Start()

	this.logger.Info("[cron job] Everyday 00:05 ETL dashboard data. entryId = %d", entryId)
}

func (this *TaskService) Bootstrap() {

	//load the clean footprint task.
	this.InitCleanFootprintTask()

	//load the etl task.
	this.InitEtlTask()

}
