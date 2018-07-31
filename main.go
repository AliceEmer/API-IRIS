package main

import (
	"github.com/AliceEmer/API-IRIS/controllers"
	"github.com/kataras/iris"

	"github.com/go-pg/pg"
)

func main() {

	db := pg.Connect(&pg.Options{
		User:     "aliceecourtemer",
		Password: "password",
		Database: "persons",
	})

	cn := &controllers.Controller{DB: db}

	app := iris.Default()

	app.Post("/signup", cn.SignUp)
	app.Post("/signin", cn.SignIn)

	//Routing group
	api := app.Party("/api", apiMiddleware)

	api.Get("/persons", cn.GetAllPersons)
	api.Get("/persons/{id:int}", cn.GetPersonByID)
	api.Post("/persons", cn.CreatePerson)
	api.Delete("/deleteperson/{id:int}", cn.DeletePerson)

	// Listen and serve on http://localhost:8080.
	app.Run(iris.Addr(":8080"))

}

func apiMiddleware(ctx iris.Context) {
	// [...]
	ctx.Next() // to move to the next handler, or don't that if you have any auth logic.
}
