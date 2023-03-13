package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type User struct {
	Id       string `json:"user"`
	Password string `json:"pass"`
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

func login(w http.ResponseWriter, req *http.Request) {
	var user User
	body, err := ioutil.ReadAll(req.Body)
	check(err)
	json.Unmarshal([]byte(body), &user)
	data, err := ioutil.ReadFile("../data/users.json")
	print(string(data))
	//w.Header().Set("Content-Type", "text/plain")
	//w.Write([]byte(usuario))
}

func main() {
	http.HandleFunc("/login", login)
	err := http.ListenAndServe("localhost:443", nil)
	check(err)
}
