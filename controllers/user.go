package controllers

import (
	"time"

	"github.com/AliceEmer/API-IRIS/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg"
	"github.com/kataras/iris"
	"golang.org/x/crypto/bcrypt"
)

//SignUp ... POST
func (cn *Controller) SignUp(c iris.Context) {

	user := models.User{}
	var usernameCheck int

	//Reading JSON data
	if err := c.ReadJSON(&user); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "creating user, read and parse form failed. " + err.Error(),
		})
		return
	}
	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if e == nil {
		user.Password = string(hashedPassword)
	}

	//Check that the needed data have been populated
	if user.Username == "" || user.Password == "" || user.Email == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Please enter a username, a password and an email to sign up",
		})
		return
	}

	//Check that the username doesn't already exist
	_, err := cn.DB.QueryOne(&usernameCheck, "SELECT id FROM users WHERE username = ?", user.Username)
	if err != nil {
		if err != pg.ErrNoRows {
			panic(err)
		}
	} else {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Username already used, please choose something else",
		})
		return
	}

	//Check that the email doesn't already exist
	_, err = cn.DB.QueryOne(&usernameCheck, "SELECT id FROM users WHERE email = ?", user.Email)
	if err != nil {
		if err != pg.ErrNoRows {
			panic(err)
		}
	} else {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "An account already exist with this email",
		})
		return
	}

	//Insertion of the new user in DB
	_, error := cn.DB.QueryOne(&user, "INSERT INTO users (username, password, email) VALUES (?, ?, ?) RETURNING * ", user.Username, user.Password, user.Email, &user)
	if error != nil {
		panic(error)
	}

	//Creation of the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"expiring": time.Now().Add(time.Hour * 72).Unix(),
	})
	user.Token, err = token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "Sorry, error while signing Token",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"token":    user.Token,
	})
}

//LogIn ... POST
func (cn *Controller) LogIn(c iris.Context) {

	user := models.User{}
	userCheck := models.User{}

	//Reading JSON data
	if err := c.ReadJSON(&user); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "creating user, read and parse form failed. " + err.Error(),
		})
		return
	}

	//Check that the needed data have been populated
	if user.Username == "" || user.Password == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Please enter a username and password to sign in",
		})
		return
	}

	//Check that the username and password exist in DB
	_, err := cn.DB.QueryOne(&userCheck, "SELECT id, username, password FROM users WHERE username = ?", user.Username)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{
				"error": "This username doesn't exist, please sign up before",
			})
			return
		}
	}

	//Password encryption
	if bcrypt.CompareHashAndPassword([]byte(userCheck.Password), []byte(user.Password)) != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Invalid username or password",
		})
		return
	}

	//Creation of the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       userCheck.ID,
		"username": userCheck.Username,
		"email":    userCheck.Email,
		"expiring": time.Now().Add(time.Hour * 72).Unix(),
	})
	user.Token, err = token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "Sorry, error while signing Token",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"OK":    "User connected",
		"token": user.Token,
	})

}
