package controllers

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/AliceEmer/API-IRIS/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg"
	mailer "github.com/kataras/go-mailer"
	"github.com/kataras/iris"
	"golang.org/x/crypto/bcrypt"
)

var (
	emailRegexp = regexp.MustCompile("^[a-z0-9._-]+@[a-z0-9._-]{2,}\\.[a-z]{2,4}$")
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

	user := models.User{}
	userCheck := models.User{}

	//Reading JSON data
	if err := c.ReadJSON(&user); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{"error": "creating user, read and parse form failed. " + err.Error()})
		return
	}

	//Check that the needed data have been populated
	if user.Username == "" || user.Password == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Please enter a username and password to login"})
		return
	}

	//Check that the username and password exist in DB
	_, err := cn.DB.QueryOne(&userCheck, "SELECT id, username, password, role FROM users WHERE username = ?", user.Username)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{"error": "This username doesn't exist, please sign up before"})
			return
		}
	}

	//Check Password
	if bcrypt.CompareHashAndPassword([]byte(userCheck.Password), []byte(user.Password+SecretSalt)) != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Invalid username or password"})
		return
	}

	//Creation of the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       userCheck.ID,
		"username": userCheck.Username,
		"role":     userCheck.Role,
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

//ValidateFormat ... Email
func ValidateFormat(email string) bool {
	if !emailRegexp.MatchString(email) {
		return false
	}
	return true
}

//UpdatePassword ... TODO: Invalid the JWT
func (cn *Controller) UpdatePassword(c iris.Context) {
	userID, _ := c.Params().GetInt("id")

	type NewPassword struct {
		Old string `json:"old,omitempty"`
		New string `json:"new,omitempty"`
	}

	newPwd := NewPassword{}
	var old string

	//Reading JSON data
	if err := c.ReadJSON(&newPwd); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "creating password, read and parse form failed. " + err.Error(),
		})
		return
	}

	//Check that the needed data have been populated
	if newPwd.Old == "" || newPwd.New == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Please enter your old and new passwords"})
		return
	}

	//Check that the old password correspond to the one of the user ID email
	_, err := cn.DB.QueryOne(&old, "SELECT password FROM users WHERE id = ?", userID)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{"error": "This user ID doesn't exist"})
			return
		}
		panic(err)

	}

	//Compare old passwords
	if bcrypt.CompareHashAndPassword([]byte(old), []byte(newPwd.Old+SecretSalt)) != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "The old password populated in not correct"})
		return
	}

	//If correct old passowrd, insertion of the new one in the DB
	hashedNewPassword, _ := bcrypt.GenerateFromPassword([]byte(newPwd.New+SecretSalt), 8)
	stringHashedNewPassword := fmt.Sprintf("%s", (hashedNewPassword))
	_, err = cn.DB.Exec("UPDATE users SET password = ? WHERE id = ?", stringHashedNewPassword, userID)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Issue updating password: " + err.Error()})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"OK": "Password updated",
	})
}

//SendValidationMail ... Sending the UUID
func (cn *Controller) SendValidationMail(c iris.Context, user *models.User) {

	config := mailer.Config{
		Host:       "smtp.ethereal.email",
		Username:   "im4h6w44cravjoss@ethereal.email",
		Password:   "cQdcnvmWFhAxaJJSUc",
		Port:       587,
		UseCommand: false,
	}

	//UUID generation
	uuid, _ := exec.Command("uuidgen").Output()
	uuidString := string(uuid)
	user.UUID = uuidString

	//Insertion of the user UUID in DB (to check it later)
	_, error := cn.DB.Exec("UPDATE users SET uuid = ? WHERE id = ?", user.UUID, user.ID)
	if error != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{"error": "Issue faced when inserting the new UUID"})
		return
	}

	//Sending the confirmation email to the user with the UUID
	sender := mailer.New(config)
	subject := "Hello subject"
	content := fmt.Sprintf("<h1>Hello %v</h1> ,<br/><br/> Please confirm you mail by copying the following code: %v  <br/><br/> See you soon ", user.Username, uuidString)
	to := []string{user.Email}

	e := sender.Send(subject, content, to...)
	if e != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{"error": "Error while sending the e-mail: " + e.Error()})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"Message":  "A confirmation email has been sent, please enter the code received below to continue.",
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})

}

//EmailVerification ...
func (cn *Controller) EmailVerification(c iris.Context) {

	userID, _ := c.Params().GetInt("id")
	user := models.User{}
	userCheck := models.User{}

	//var UUID string

	//Check that the user exist in DB
	_, err := cn.DB.QueryOne(&user, "SELECT * FROM users WHERE id = ?", userID)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{"error": "This user doesn't exist"})
			return
		}
	}

	//Reading UUID JSON data
	if err := c.ReadJSON(&userCheck); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "creating UUID, read and parse form failed. " + err.Error(),
		})
		return
	}

	//Check that the needed data have been populated
	if userCheck.UUID == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Please enter the code received by email",
		})
		return
	}

	//Check that the UUID are the same
	if userCheck.UUID != strings.TrimRight(user.UUID, "\n") {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "The code populated is not the one that has been sent to you by email",
		})
		return
	}

	//If they are the same, we pass the email_validated at TRUE in the DB, and the UUID to empty
	_, err = cn.DB.Exec("UPDATE users SET email_validated = true, UUID = null WHERE id = ?", userID)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Issue updating email_validated: " + err.Error()})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"OK": "Email validated",
	})

}
