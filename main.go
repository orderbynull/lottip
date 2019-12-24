package main

import (
	app2 "github.com/orderbynull/lottip/core"
	"github.com/orderbynull/lottip/impl"
	"github.com/orderbynull/lottip/web"
)

func main() {
	pgsqlService := app2.NewPgsqlService(&impl.MemoryPgsqlRepository{})

	app := &web.UiApp{PgsqlService: pgsqlService, PageSize: 3}

	app.AddRouteHandler("/pgsql", web.PostgresqlHandler)

	app.Run(":8081")
}
