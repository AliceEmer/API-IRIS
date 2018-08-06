package controllers

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/AliceEmer/API-IRIS/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgryski/dgoogauth"
	"github.com/go-pg/pg"
	"github.com/kataras/iris"
	"golang.org/x/crypto/bcrypt"
)

var (
	emailRegexp = regexp.MustCompile("^[a-z0-9._-]+@[a-z0-9._-]{2,}\\.[a-z]{2,4}$")
)

//ValidateFormat ... Email
func ValidateFormat(email string) bool {
	if !emailRegexp.MatchString(email) {
		return false
	}
	return true
}

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

	//Check that the email has a correct format
	isValid := ValidateFormat(user.Email)
	if isValid != true {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Your email doesn't have a correct format"})
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
		c.JSON(iris.Map{"error": "An account already exist with this email"})
		return
	}

	//If everything is ok, hashing the password
	saltedPassword := user.Password + SecretSalt
	hashedPassword, e := bcrypt.GenerateFromPassword([]byte(saltedPassword), 8)
	if e == nil {
		user.Password = string(hashedPassword)
	}

	//Insertion of the new user in DB
	_, error := cn.DB.QueryOne(&user, "INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?) RETURNING * ", user.Username, user.Password, user.Email, user.Role, &user)
	if error != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{"error": "Issue faced when insertin the new user"})
		return
	}

	//Send the email to the user
	cn.SendValidationMail(c, &user)

}

//LogIn ... POST
func (cn *Controller) LogIn(c iris.Context) {

	userLogIn := models.User{}
	user := models.User{}

	//Reading JSON data
	if err := c.ReadJSON(&userLogIn); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{"error": "creating user, read and parse form failed. " + err.Error()})
		return
	}

	//Check that the needed data have been populated
	if userLogIn.Username == "" || userLogIn.Password == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Please enter a username and password to login"})
		return
	}

	//Check that the username and password exist in DB
	_, err := cn.DB.QueryOne(&user, "SELECT * FROM users WHERE username = ?", userLogIn.Username)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{"error": "This username doesn't exist, please sign up before"})
			return
		}
	}

	//Check Password
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userLogIn.Password+SecretSalt)) != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Invalid username or password"})
		return
	}

	if user.Twofa_activated == true {
		//Check that the 2fa token is correct
		if userLogIn.TwoFA_token == "" {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{"error": "Please enter your Two Factor Authentification Token"})
			return
		}
		for cn.LogIn2FA(c, userLogIn.TwoFA_token) == false {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{"error": "The token provided is not correct, please try again"})
		}
	}

	//Creation of the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		"expiring": time.Now().Add(time.Hour * 72).Unix(),
	})
	user.Token, err = token.SignedString([]byte(JWTSecretKey))
	if err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{"error": "Sorry, error while signing Token"})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"OK":    "User connected",
		"token": user.Token,
	})
}

//LogIn2FA ... POST -- Not tested as the QR code generated is not valid on Google Auth
func (cn *Controller) LogIn2FA(c iris.Context, token string) bool {

	// setup the one-time-password configuration.
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      strings.TrimSpace(TWOFASecretKey),
		WindowSize:  3,
		HotpCounter: 0,
	}

	trimmedToken := strings.TrimSpace(token)

	// Validate token
	ok, err := otpConfig.Authenticate(trimmedToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return ok

}
