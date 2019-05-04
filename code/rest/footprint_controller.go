package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"net/http"
)

type FootprintController struct {
	BaseController
	footprintDao     *FootprintDao
	footprintService *FootprintService
}

func (this *FootprintController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.footprintDao)
	if b, ok := b.(*FootprintDao); ok {
		this.footprintDao = b
	}

	b = core.CONTEXT.GetBean(this.footprintService)
	if b, ok := b.(*FootprintService); ok {
		this.footprintService = b
	}

}

func (this *FootprintController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	return routeMap
}
