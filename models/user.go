package models

//User ...
type User struct {
	ID       string `json:"id,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
	Token    string `json:"jwt,omitempty"`
}
