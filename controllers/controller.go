package controllers

import "github.com/go-pg/pg"

//Controller ... Database connection
type Controller struct {
	DB *pg.DB
}
