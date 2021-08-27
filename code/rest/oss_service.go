package rest

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/eyebluecn/tank/code/core"
)

// oss service
//@Service
type OssService struct {
	BaseBean
	footprintService *FootprintService

	//whether scan task is running
	client *oss.Client
}

func (this *OssService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.footprintService)
	if b, ok := b.(*FootprintService); ok {
		this.footprintService = b
	}

}

//init the elt task.
func (this *OssService) InitEtlTask() {

	this.logger.Info("[cron job] Everyday 00:05 ETL dashboard data.")
}

func (this *OssService) Bootstrap() {

	// Endpoint以杭州为例，其它Region请按实际情况填写。
	// 阿里云主账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM账号进行API访问或日常运维，请登录 https://ram.console.aliyun.com 创建RAM账号。

	// 创建OSSClient实例。
	client, err := oss.New(fmt.Sprintf("http://%s", core.CONFIG.OssAccessKey()), core.CONFIG.OssAccessKey(), core.CONFIG.OssSecretKey())
	this.client = client
	if err != nil {
		panic(err)
	}

	this.logger.Info("OssService bootstrap.")

}
