package main

import (
	"github.com/AliceEmer/API-IRIS/controllers"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"

	"github.com/go-pg/pg"
)

var cn = &controllers.Controller{}

func main() {

	//DB connection
	db := pg.Connect(&pg.Options{
		User:     "aliceecourtemer",
		Password: "password",
		Database: "persons",
	})
	defer db.Close()
	cn.DB = db

	app := iris.New()

	//CORS middleware
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"Get", "Post", "Delete", "Put"},
	})

	//Authentification
	app.Post("/signup", cn.SignUp, crs)
	app.Post("/signin", cn.SignIn, crs)

	//Routing group api
	api := app.Party("/api", apiMiddleware, crs)

	api.Get("/persons", cn.GetAllPersons)
	api.Get("/person/{id:int}", cn.GetPersonByID)
	api.Get("/person/{id:int}/address", cn.GetAddressByPerson)

	api.Post("/addperson", cn.CreatePerson)
	api.Post("/addaddress/{id:int}", cn.CreateAddress)

	api.Delete("/deleteperson/{id:int}", cn.DeletePerson)
	api.Delete("/deleteaddress/{id:int}", cn.DeleteAddress)

	// Listen and serve on http://localhost:8080.
	app.Run(iris.Addr(":8080"))

}

//No access to api/ if the JWT is not created or not valid
func apiMiddleware(c iris.Context) {
	if cn.CheckJWT(c) {
		c.Next()
	}
}
