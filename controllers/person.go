package controllers

import (
	"github.com/AliceEmer/API-IRIS/models"
	"github.com/kataras/iris"
)

//GetAllPersons ... GET
func (cn *Controller) GetAllPersons(c iris.Context) {
	var persons []models.Person

	_, err := cn.DB.Query(&persons, "SELECT * FROM person")
	if err != nil {
		c.Values().Set("error", "Selecting persons failed. "+err.Error())
		c.StatusCode(iris.StatusInternalServerError)
		return
	}

	if len(persons) == 0 {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "No person in the databse",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"people": persons,
	})
}

//GetPersonByID ... GET
func (cn *Controller) GetPersonByID(c iris.Context) {

	personID, _ := c.Params().GetInt("id")

	var person models.Person

	_, err := cn.DB.QueryOne(&person, "SELECT * FROM person WHERE id = ?", personID)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "No person with this ID in the databse",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"people": person,
	})

}

//CreatePerson ... POST
func (cn *Controller) CreatePerson(c iris.Context) {

	person := models.Person{}

	if err := c.ReadJSON(&person); err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.Values().Set("error", "creating user, read and parse form failed. "+err.Error())
		return
	}
	_, err := cn.DB.QueryOne(&person, "INSERT INTO person VALUES (?, ?) RETURNING id ", person.Firstname, person.Lastname, &person)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"id":        person.ID,
		"firstname": person.Firstname,
		"Lastname":  person.Lastname,
	})

}

//DeletePerson ... DELETE
func (cn *Controller) DeletePerson(c iris.Context) {

	personID, _ := c.Params().GetInt("id")

	_, err := cn.DB.Exec("DELETE FROM person WHERE id = ? RETURNING * ", personID)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"message": "Person deleted",
	})

}
