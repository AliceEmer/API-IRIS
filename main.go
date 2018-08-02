package main

import (
	"github.com/AliceEmer/API-IRIS/controllers"
	jwt "github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"

	"github.com/go-pg/pg"
)

//DB connexion variables
const (
	UserDB     = "aliceecourtemer"
	PasswordDB = "password"
	NameDB     = "persons"
)

func main() {

	//DB connection
	db := pg.Connect(&pg.Options{
		User:     UserDB,
		Password: PasswordDB,
		Database: NameDB,
	})
	defer db.Close()

	//Controller holding DB connection
	var cn = &controllers.Controller{}
	cn.DB = db

	app := iris.New()

	//CORS middleware
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"Get", "Post", "Delete", "Put"},
	})
	app.Use(crs)

	//JWT middleware
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(controllers.JWTSecretKey), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	//AdminCheck middleware
	adminMiddleware := func(c iris.Context) {
		userToken := jwtMiddleware.Get(c)
		claims, _ := userToken.Claims.(jwt.MapClaims)
		if claims["role"].(float64) == 0 {
			c.StatusCode(iris.StatusUnauthorized)
			c.JSON(iris.Map{"error": "You don't have enough rights to access to this page"})
			return
		}
		c.Next()
	}

	//Authentification
	app.Post("/signup", cn.SignUp)
	app.Post("/login", cn.LogIn)

	//Routing group api
	api := app.Party("/api")
	api.Use(jwtMiddleware.Serve)

	api.Get("/persons", cn.GetAllPersons)
	api.Get("/person/{id:int}", cn.GetPersonByID)
	api.Get("/person/{id:int}/address", cn.GetAddressByPerson)

	api.Post("/addperson", cn.CreatePerson)
	api.Post("/addaddress/{id:int}", cn.CreateAddress)

	//Routing group admin
	admin := api.Party("/admin")
	admin.Use(adminMiddleware)

	admin.Delete("/deleteperson/{id:int}", cn.DeletePerson)
	admin.Delete("/deleteaddress/{id:int}", cn.DeleteAddress)

	// Listen and serve on http://localhost:8080.
	app.Run(iris.Addr(":8080"))

}
