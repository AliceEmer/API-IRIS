package models

//User ...
type User struct {
	ID             int    `json:"id,omitempty"`
	Password       string `json:"password,omitempty"`
	Username       string `json:"username,omitempty"`
	Email          string `json:"email,omitempty"`
	Token          string `json:"token,omitempty"`
	Role           int    `json:"role,omitempty"`
	UUID           string `json:"uuid,omitempty"`
	EmailValidated bool   `json:"email_validated,omitempty"`
}
