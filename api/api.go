package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
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

func register(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	//Read body and save in User Struct
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	json.Unmarshal([]byte(body), &user)

	// Read users.json and map
	data, fileErr := os.ReadFile("../data/users.json")
	check(fileErr)
	users := make(map[string]string)
	json.Unmarshal([]byte(data), &users)

	if _, ok := users[user.Id]; ok {
		resp := make(map[string]string)
		resp["msg"] = "User already exists"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	} else {
		users[user.Id] = user.Password
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		check(erro)
		resp := make(map[string]string)
		resp["msg"] = "User correctly registered"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	}
}

func main() {
	http.HandleFunc("/register", register)
	err := http.ListenAndServe("localhost:443", nil)
	check(err)
}
