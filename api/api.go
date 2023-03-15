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

func mailExists(users map[string]string, user User, needPassword bool) bool {
	pswd, ok := users[user.Id]
	if needPassword {
		ok := user.Password == pswd
		return ok
	}
	return ok
}

// Falta guardar las contraseñas hasheadas
func register(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	//Read body and save in User Struct
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	json.Unmarshal(body, &user)

	// Read users.json and map
	data, fileErr := os.ReadFile("../data/users.json")
	check(fileErr)
	users := make(map[string]string)
	json.Unmarshal(data, &users)

	if mailExists(users, user, false) {
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

// Maps users.json
func mapUserJson() map[string]string {
	data, fileErr := os.ReadFile("../data/users.json")
	check(fileErr)
	users := make(map[string]string)
	json.Unmarshal(data, &users)
	return users
}

// Assigns body parameters to a user
func getBodyUser(req *http.Request) User {
	var user User
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	json.Unmarshal(body, &user)
	return user
}

// Falta adjuntar la clave
func login(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	user = getBodyUser(req)
	users := mapUserJson()
	/*
		hash := sha256.New()
		hash.Write([]byte(user.Password))
	*/
	if mailExists(users, user, true) {
		resp := make(map[string]string)
		resp["msg"] = "User correctly logged"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "Incorrect user or password"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

func main() {
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	err := http.ListenAndServe("localhost:443", nil)
	check(err)
}
