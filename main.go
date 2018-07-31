package main

import (
	"github.com/AliceEmer/API-IRIS/controllers"
	"github.com/kataras/iris"

	"github.com/go-pg/pg"
)

var cn = &controllers.Controller{}

func main() {

	db := pg.Connect(&pg.Options{
		User:     "aliceecourtemer",
		Password: "password",
		Database: "persons",
	})

	cn.DB = db

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

//No access to api/ if the JWT is not created or not valid
func apiMiddleware(c iris.Context) {
	// [...]
	if cn.CheckJWT(c) {
		c.Next()
	}
}
