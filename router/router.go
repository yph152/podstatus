package router

import (
	//"fmt"
	"github.com/emicklei/go-restful"
	"net/http"
)

var webService *restful.WebService

var wsContainer *restful.Container

func InitRouterService(prefix string) {
	wsContainer = restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})

	webService = newWebService(wsContainer, prefix)

}
func newWebService(container *restful.Container, prefix string) *restful.WebService {
	ws := new(restful.WebService)

	ws.Path(prefix).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	container.Add(ws)
	return ws
}
func Get() *restful.WebService {
	return webService
}

func Run() {
	server := &http.Server{Addr: ":8088", Handler: wsContainer}

	server.ListenAndServe()
}
