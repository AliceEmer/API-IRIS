package controllers

import (
	"github.com/AliceEmer/API-IRIS/models"
	"github.com/kataras/iris"
)

//GetAllPersons ... GET
func (cn *Controller) GetAllPersons(c iris.Context) {

	rows, err := cn.DB.Query("SELECT * FROM person")
	if err != nil {
		c.Values().Set("error", "Selecting persons failed. "+err.Error())
		c.StatusCode(iris.StatusInternalServerError)
		return
	}
	defer rows.Close()

	pers := make([]*models.Person, 0)
	for rows.Next() {
		p := new(models.Person)
		err := rows.Scan(&p.Firstname, &p.Lastname, &p.ID)
		if err != nil {
			c.StatusCode(iris.StatusBadRequest)
			c.WriteString(err.Error())
			return
		}
		pers = append(pers, p)
	}
	if err = rows.Err(); err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.WriteString(err.Error())
		return
	}

	if len(pers) == 0 {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "No person in the databse",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"people": pers,
	})
}

//GetPersonByID ... GET
func (cn *Controller) GetPersonByID(c iris.Context) {

	personID, _ := c.Params().GetInt("id")

	rows, err := cn.DB.Query("SELECT * FROM person WHERE id = $1", personID)
	if err != nil {
		c.Values().Set("error", "Selecting persons failed. "+err.Error())
		c.StatusCode(iris.StatusInternalServerError)
		return
	}
	defer rows.Close()

	pers := make([]*models.Person, 0)
	for rows.Next() {
		p := new(models.Person)
		err := rows.Scan(&p.Firstname, &p.Lastname, &p.ID)
		if err != nil {
			c.StatusCode(iris.StatusBadRequest)
			c.WriteString(err.Error())
			return
		}
		pers = append(pers, p)
	}

	if err = rows.Err(); err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.WriteString(err.Error())
		return
	}

	if len(pers) == 0 {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(iris.Map{
			"error": "No person in the databse",
		})
		return
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"people": pers,
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

	_, err := cn.DB.Exec("INSERT INTO person VALUES ($1, $2)", person.Firstname, person.Lastname)
	if err != nil {
		panic(err)
	}

	c.StatusCode(iris.StatusOK)
	c.JSON(iris.Map{
		"firstname": person.Firstname,
		"Lastname":  person.Lastname,
	})

}
