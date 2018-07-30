package controllers

import (
	"github.com/AliceEmer/API-IRIS/models"
	"github.com/go-pg/pg"
	"github.com/kataras/iris"
	"golang.org/x/crypto/bcrypt"
)

//SignUp ... POST
func (cn *Controller) SignUp(c iris.Context) {

	user := models.User{}

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

	_, err := cn.DB.QueryOne(&user, "INSERT INTO users (username, password) VALUES (?, ?) RETURNING * ", user.Username, user.Password, &user)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"id":       user.ID,
		"username": user.Username,
	})
}

//SignIn ... POST
func (cn *Controller) SignIn(c iris.Context) {

	user := models.User{}
	userCheck := models.User{}

	if err := c.ReadJSON(&user); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "creating user, read and parse form failed. " + err.Error(),
		})
		return
	}

	_, err := cn.DB.QueryOne(&userCheck, "SELECT username, password FROM users WHERE username = ?", user.Username)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{
				"error": "This username doesn't exist, please sign up before",
			})
			return
		}
	}

	if bcrypt.CompareHashAndPassword([]byte(userCheck.Password), []byte(user.Password)) != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Invalid username or password",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"OK":       "User connected",
		"username": user.Username,
	})
}
