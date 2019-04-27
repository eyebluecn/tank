package core

import (
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/jinzhu/gorm"
	"net/http"
)

type Context interface {
	//具备响应http请求的能力
	http.Handler

	//获取数据库链接
	GetDB() *gorm.DB

	//获取一个Bean
	GetBean(bean Bean) Bean

	//获取全局的Session缓存
	GetSessionCache() *cache.Table

	//获取全局的ControllerMap
	GetControllerMap() map[string]Controller

	//系统安装成功
	InstallOk()

	//清空系统
	Cleanup()
}
