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
	matterDao *MatterDao

	//whether scan task is running
	client *oss.Client
	bucket *oss.Bucket
}

func (this *OssService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

}

//upload a matter to oss.
func (this *OssService) Upload(matter *Matter) {

	//只上传文件，忽略文件夹。
	if matter.Dir {
		return
	}

	//只上传没有传过的。
	if matter.Md5 != "" {
		return
	}

	// <yourObjectName>上传文件到OSS时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
	objectName := matter.Path
	if len(matter.Path) > 1 {
		objectName = matter.Path[1:]
	}

	// <yourLocalFileName>由本地文件路径加文件名包括后缀组成，例如/users/local/myfile.txt。
	localFileName := matter.AbsolutePath()

	// 上传文件。
	err := this.bucket.PutObjectFromFile(objectName, localFileName)
	this.PanicError(err)

	// 修改本地文件的md5字段
	matter.Md5 = fmt.Sprintf("%s/%s", core.CONFIG.OssCustomHost(), objectName)
	this.matterDao.Save(matter)

	this.logger.Info("upload %s %s to oss path = %s", matter.Uuid, matter.Name, objectName)

}

func (this *OssService) Bootstrap() {

	// Endpoint以杭州为例，其它Region请按实际情况填写。
	// 阿里云主账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM账号进行API访问或日常运维，请登录 https://ram.console.aliyun.com 创建RAM账号。

	// 创建OSSClient实例。
	client, err := oss.New(fmt.Sprintf("http://%s", core.CONFIG.OssEndpoint()), core.CONFIG.OssAccessKey(), core.CONFIG.OssSecretKey())
	this.client = client
	this.PanicError(err)

	// 获取存储空间。
	bucket, err := this.client.Bucket(core.CONFIG.OssBucket())
	this.bucket = bucket
	this.PanicError(err)

	this.logger.Info("OssService bootstrap.")

}
