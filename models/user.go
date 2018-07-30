package models

//User ...
type User struct {
	ID       string `json:"id"`
	Password string `json:"password"`
	Username string `json:"username"`
}
