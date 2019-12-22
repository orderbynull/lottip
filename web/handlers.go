package web

import (
	"html/template"
	"log"
	"net/http"
)

func PostgresqlHandler(app *UiApp, w http.ResponseWriter, r *http.Request) {
	rows := app.PgsqlService.GetAllByApp("")

	tmpl := template.Must(template.ParseFiles("./web/templates/layout.html", "./web/templates/pgsql.html"))
	if err := tmpl.ExecuteTemplate(w, "layout", rows); err != nil {
		log.Println(err)
	}
}
