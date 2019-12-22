package web

import (
	"github.com/orderbynull/lottip/app"
	"log"
	"net/http"
)

type RouteHandler func(app *UiApp, w http.ResponseWriter, r *http.Request)

type UiApp struct {
	PgsqlService *app.PgsqlService
}

func (app *UiApp) withApp(fn RouteHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		fn(app, w, r)
	}
}

func (app *UiApp) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (app *UiApp) AddRouteHandler(route string, handler RouteHandler) {
	http.HandleFunc(route, app.withApp(handler))
}
