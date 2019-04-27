package core

import (
	"github.com/eyebluecn/tank/code/tool/cache"
	"github.com/jinzhu/gorm"
	"net/http"
)

type Context interface {
	//获取数据库链接
	GetDB() *gorm.DB

	//获取一个Bean
	GetBean(bean IBean) IBean

	//获取全局的Session缓存
	GetSessionCache() *cache.Table

	//获取全局的ControllerMap
	GetControllerMap() map[string]IController

	//响应http的能力
	ServeHTTP(writer http.ResponseWriter, request *http.Request)

	//系统安装成功
	InstallOk()

	//清空系统
	Cleanup()
}
