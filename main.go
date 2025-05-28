package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Person struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Age  int    `db:"age" json:"age"`
}

var db *sqlx.DB

func main() {
	// Подключение к БД
	var err error
	dsn := "host=localhost port=5433 user=root password=docker dbname=medtest sslmode=disable"
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		panic(err)
	}

	// Создаем таблицу, если не существует
	schema := `CREATE TABLE IF NOT EXISTS peoples (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100),
  age INT
 );`
	db.MustExec(schema)

	router := gin.Default()

	router.POST("/peoples", createPerson)
	router.GET("/peoples", getAllPeople)
	router.GET("/people/:id", getPersonByID)

	router.Run(":8080")
}

func createPerson(c *gin.Context) {
	log.Println("POST /peoples — получен запрос на создание человека")
	var p Person
	if err := c.BindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p.Age <= 0 {
		log.Println("Возраст не может быть отрицательным")
		return
	}
	query := `INSERT INTO peoples(name, age) VALUES($1, $2) RETURNING id`
	err := db.QueryRow(query, p.Name, p.Age).Scan(&p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func getAllPeople(c *gin.Context) {
	log.Println("GET /peoples — получен запрос на получение всех людей")
	var peoples []Person
	err := db.Select(&peoples, "SELECT id, name, age FROM peoples")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, peoples)
}

func getPersonByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	log.Println("получен запрос на получение человека по ID", id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var p Person
	err = db.Get(&p, "SELECT id, name, age FROM peoples WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}
