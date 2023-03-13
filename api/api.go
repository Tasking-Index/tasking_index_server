package main

import (
	"github.com/gin-gonic/gin"
)

type user struct {
	Id       string
	Password string
}

type project struct {
	Id          int
	Titulo      string
	Descripcion string
	Archivos    []string
}

type additional struct {
	fecha_inicio  string
	fecha_fin     string
	propietario   string
	colaboradores []string
	tareas        []tarea
	logs          []string
}

type tarea struct {
	nombre string
	estado bool
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func test(c *gin.Context) {
	print("Hola")
}

func main() {
	router := gin.Default()
	router.GET("/", test)
	router.Run("localhost:8080")
}
