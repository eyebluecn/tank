package core

import "net/http"

type Controller interface {
	Bean
	//register self's fixed routes
	RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request)
	//handle some special routes, eg. params in the url.
	HandleRoutes(writer http.ResponseWriter, request *http.Request) (func(writer http.ResponseWriter, request *http.Request), bool)
}
