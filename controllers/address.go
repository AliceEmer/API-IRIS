package controllers

import (
	"github.com/AliceEmer/API-IRIS/models"
	"github.com/go-pg/pg"
	"github.com/kataras/iris"
)

//GetAddressByPerson ... GET
func (cn *Controller) GetAddressByPerson(c iris.Context) {

	personID, _ := c.Params().GetInt("id")

	//Check that the person_id correspond to a person in DB
	var person models.Person
	_, e := cn.DB.QueryOne(&person, "SELECT * FROM persons WHERE id = ?", personID)
	if e != nil {
		if e == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{
				"error": "The person populated doesn't exist. Impossible to add address",
			})
			return
		}
		panic(e)
	}

	var address []models.Address
	_, err := cn.DB.Query(&address, "SELECT * FROM addresses WHERE person_id = ?", personID)
	if err != nil {
		if err == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{
				"error": "No address for a person with this ID in the database",
			})
			return
		}
		panic(err)
	}

	//Check that the address map is not empty
	if len(address) == 0 {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "No address for a person with this ID in the database",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"address": address,
		"person":  person,
	})
}

//CreateAddress ... POST
func (cn *Controller) CreateAddress(c iris.Context) {

	personID, _ := c.Params().GetInt("id")

	//Check that the person_id correspond to a person in DB
	var person models.Person
	_, e := cn.DB.QueryOne(&person, "SELECT * FROM persons WHERE id = ?", personID)
	if e != nil {
		if e == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{
				"error": "The person populated doesn't exist. Impossible to add address",
			})
			return
		}
		panic(e)
	}

	//Reading JSON data
	var address models.Address
	if err := c.ReadJSON(&address); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.Values().Set("error", "Creating address, read and parse form failed. "+err.Error())
		return
	}

	//Check that the needed data have been populated
	if address.City == "" || address.State == "" {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "Please enter the city and state of the new address",
		})
		return
	}

	_, err := cn.DB.QueryOne(&address, "INSERT INTO addresses(city, state, person_id)  VALUES (?, ?, ?) RETURNING * ", address.City, address.State, personID, &address)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"person":  person,
		"address": address,
	})

}

//DeleteAddress ... DELETE
func (cn *Controller) DeleteAddress(c iris.Context) {

	personID, _ := c.Params().GetInt("id")

	//Check that the person_id correspond to a person in DB
	var person models.Person
	_, e := cn.DB.QueryOne(&person, "SELECT * FROM persons WHERE id = ?", personID)
	if e != nil {
		if e == pg.ErrNoRows {
			c.StatusCode(iris.StatusBadRequest)
			c.JSON(iris.Map{
				"error": "The person populated doesn't exist. Impossible to delete the address",
			})
			return
		}
		panic(e)
	}

	_, err := cn.DB.Exec("DELETE FROM addresses WHERE person_id = ?", personID)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"message": "Address deleted",
	})
}
