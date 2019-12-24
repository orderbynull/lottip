package web

import (
	"github.com/orderbynull/lottip/core"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func PostgresqlHandler(uiApp *UiApp, w http.ResponseWriter, r *http.Request) {
	newer := r.URL.Query().Get("newer")
	older := r.URL.Query().Get("older")
	app := r.URL.Query().Get("app")

	var rows []core.PgsqlPacket
	if newer != "" {
		val, _ := strconv.Atoi(newer)
		rows = uiApp.PgsqlService.GetAllByApp(app, "newer", val, uiApp.PageSize)
	} else {
		val, _ := strconv.Atoi(older)
		rows = uiApp.PgsqlService.GetAllByApp(app, "older", val, uiApp.PageSize)
	}

	data := struct {
		Rows      []core.PgsqlPacket
		Paginator *core.Paginator
	}{
		rows,
		core.NewPaginator(r, rows),
	}

	tmpl := template.Must(template.ParseFiles("./web/templates/layout.html", "./web/templates/pgsql.html"))
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Println(err)
	}
}
