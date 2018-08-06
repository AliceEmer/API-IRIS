package controllers

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"image"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/AliceEmer/API-IRIS/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/disintegration/imaging"
	"github.com/go-pg/pg"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
	mailer "github.com/kataras/go-mailer"
	"github.com/kataras/iris"
	"golang.org/x/crypto/bcrypt"
	"rsc.io/qr"
)

//UpdatePassword ...
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

	//Invalid token
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(JWTSecretKey), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	userToken := jwtMiddleware.Get(c)
	claims, _ := userToken.Claims.(jwt.MapClaims)
	claims["expiring"] = time.Now() //not sure if it's enough

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"OK": "Password updated, you need to login again with your new password",
	})
}

//UpdateRole ...
func (cn *Controller) UpdateRole(c iris.Context) {
	userID, _ := c.Params().GetInt("id")
	user := models.User{}

	//Reading JSON data
	if err := c.ReadJSON(&user); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(iris.Map{
			"error": "creating role, read and parse form failed. " + err.Error(),
		})
		return
	}

	//Insertion of the new role in DB
	_, err := cn.DB.Exec("UPDATE users SET role = ? WHERE id = ?", user.Role, userID)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Issue updating password: " + err.Error()})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{"OK": "Role updated"})
}

//DeleteUser ... DELETE
func (cn *Controller) DeleteUser(c iris.Context) {

	userID, _ := c.Params().GetInt("id")
	_, err := cn.DB.Exec("DELETE FROM users WHERE id = ? RETURNING * ", userID)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{"message": "User deleted"})
}

//SendValidationMail ... Sending the UUID - Using Ethereal for now
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

//Activate2FA ... 2FA activation --> TODO : confirmation with the password, login again at the end
func (cn *Controller) Activate2FA(c iris.Context) {
	userID, _ := c.Params().GetInt("id")

	// maximize CPU usage for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	// generate a random string - preferably 6 or 8 characters
	secret := base32.StdEncoding.EncodeToString([]byte(TWOFASecretKey))
	authLink := "otpauth://totp/Hades?secret=" + secret + "&issuer=Hades"

	code, err := qr.Encode(authLink, qr.L) //L is the lowest lovel of error correction level (7% only can be damaged)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Error creating the QR code: " + err.Error()})
		return
	}

	// convert byte to image for saving to file
	imgByte := code.PNG()
	img, _, _ := image.Decode(bytes.NewReader(imgByte))

	err = imaging.Save(img, "./QRImgHades.png")
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Error saving the QR code image: " + err.Error()})
		return
	}

	//Two Factor Auth set to true in DB for this user
	_, e := cn.DB.Exec("UPDATE users SET twofa_activated = true WHERE id = ?", userID)
	if e != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{"error": "Issue updating 2FA activation: " + e.Error()})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{"OK": "QR code generated and saved to QRImgHades.png in the current repository."})

	//c.Redirect("/login")
	// ---> Redirection to LOGIN

}
