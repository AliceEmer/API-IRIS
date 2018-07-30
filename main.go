package main

import (
	"database/sql"
	"log"

	"github.com/AliceEmer/API-IRIS/controllers"
	"github.com/kataras/iris"

	_ "github.com/lib/pq"
)

func main() {

	database := InitDB("postgres://aliceecourtemer:password@localhost/persons?sslmode=disable")

	cn := &controllers.Controller{DB: database}

	app := iris.Default()

	//api := app.Party("/api")

	app.Get("/persons", cn.GetAllPersons)
	app.Get("/persons/{id:int}", cn.GetPersonByID)

	app.Post("/persons", cn.CreatePerson)

	// Listen and serve on http://localhost:8080.
	app.Run(iris.Addr(":8080"))

}

//InitDB ...
func InitDB(dataSourceName string) *sql.DB {
	var db *sql.DB
	var err error

	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Panic(err)
	}

	//defer db.Close()

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	return db
}
