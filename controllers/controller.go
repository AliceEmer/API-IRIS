package controllers

import (
	"database/sql"
)

//Controller ... Database connection
type Controller struct {
	DB *sql.DB
}
