package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		tmpl := template.Must(template.ParseFiles("ui/templates/layout.html", "ui/templates/default.html"))
		if err := tmpl.ExecuteTemplate(writer, "layout", nil); err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/mysql", func(writer http.ResponseWriter, request *http.Request) {
		tmpl := template.Must(template.ParseFiles("ui/templates/layout.html", "ui/templates/mysql.html"))
		if err := tmpl.ExecuteTemplate(writer, "layout", nil); err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/pgsql", func(writer http.ResponseWriter, request *http.Request) {
		tmpl := template.Must(template.ParseFiles("ui/templates/layout.html", "ui/templates/pgsql.html"))
		if err := tmpl.ExecuteTemplate(writer, "layout", nil); err != nil {
			log.Println(err)
		}
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
