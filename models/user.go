package models

//User ...
type User struct {
	ID       string `json:"id,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
	Role     int    `json:"role,omitempty"`
}
