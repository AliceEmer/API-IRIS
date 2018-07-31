package controllers

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg"
	"github.com/kataras/iris"
)

//Controller ... Database connection
type Controller struct {
	DB  *pg.DB
	JWT string
}

//JWTSecretKey ... Signing Key for JWT
const (
	JWTSecretKey = "SigningKey"
)

//CheckJWT ... Veryfying that the JWT is valid
func (cn *Controller) CheckJWT(c iris.Context) bool {

	if cn.JWT == "" {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "Sorry, you need to login first",
		})
		return false

	}
	token, err := jwt.Parse(cn.JWT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecretKey), nil
	})
	if err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "Unexpected issue parsing your token",
		})
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Printf("\nid : %v, user : %v \n", claims["id"], claims["username"])
		return true
	}

	return false
}
