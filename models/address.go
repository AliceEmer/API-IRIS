package models

//Address ...
type Address struct {
	ID       string `json:"id,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	PersonID string `json:"person_id,omitempty"`
}
