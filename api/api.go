package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       string `json:"user"`
	Password string `json:"pass"`
}

type UsersProjects struct {
	Users    map[string][]int `json:"users"`
	Projects []Project        `json:"projects"`
}

type Project struct {
	Id          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
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

func userExists(users map[string]string, user User, needPassword bool) bool {
	pswd, ok := users[user.Id]
	if needPassword {
		//ok := user.Password == pswd
		ok := comparePasswords(user.Password, []byte(pswd))
		return ok
	}
	return ok
}

func hash(s string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	return hasher.Sum(nil)
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	print(hash)
	print("\n")
	print(string(hash))

	return string(hash)
}

func comparePasswords(hashedPwd string, bytePwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(bytePwd, byteHash)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Maps users.json
func mapUserJson() map[string]string {
	data, fileErr := os.ReadFile("../data/users.json")
	check(fileErr)
	users := make(map[string]string)
	error := json.Unmarshal(data, &users)
	check(error)
	return users
}

func getProjectsJson() UsersProjects {
	data, fileErr := os.ReadFile("../data/projects.json")
	check(fileErr)
	var projects UsersProjects
	error := json.Unmarshal(data, &projects)
	check(error)
	return projects
}

// Assigns body parameters to a user
func getBodyUser(req *http.Request) User {
	var user User
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	error := json.Unmarshal(body, &user)
	check(error)
	return user
}

// TODO
func projectsMatched(userId string) {
	//Save json to struct
	var usersProjects UsersProjects
	data, fileErr := os.ReadFile("../data/projects.json")
	check(fileErr)
	json.Unmarshal(data, &usersProjects)
}

// Request: {"user":"user1", "pass":"pass1", "title": "Title Project"}
// TODO
func createProject(w http.ResponseWriter, req *http.Request) {
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	users := mapUserJson()
	var user User
	json.Unmarshal(body, &user)
	if userExists(users, user, true) {
		bodyInfo := make(map[string]string)
		json.Unmarshal(body, &bodyInfo)
		data, fileErr := os.ReadFile("../data/projects.json")
		check(fileErr)
		var usersProjects UsersProjects
		var project Project
		json.Unmarshal(data, &usersProjects)
		lastItem := len(usersProjects.Projects)
		if lastItem == 0 {
			project.Id = 0
		} else {
			project.Id = usersProjects.Projects[lastItem-1].Id + 1
		}
		project.Title = bodyInfo["title"]
		project.Description = ""
		project.Files = nil
		usersProjects.Projects = append(usersProjects.Projects, project)
		usersProjectsJSON, JsonErr := json.MarshalIndent(usersProjects, "", "  ")
		check(JsonErr)
		erro := os.WriteFile("../data/projects.json", usersProjectsJSON, 0666)
		check(erro)
		resp := make(map[string]interface{})
		resp["msg"] = "Project correctly created"
		resp["id"] = project.Id
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "Login needed to create a project"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(400)
		w.Write(jsonResp)
	}
}

// Request {"id": 1, "user":"paco", "pass": "paco1"}
func getProject(w http.ResponseWriter, req *http.Request) {
	print("HOLA")
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	users := mapUserJson()
	var user User
	json.Unmarshal(body, &user)
	if userExists(users, user, true) {
		//DAR INFORMACIÓN DEL PROYECTO
		print("EXISTE")
		projects := getProjectsJson()
		fmt.Println(projects)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "Login needed to get projects"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(400)
		w.Write(jsonResp)
	}
}

// TODO añadir usuario a UsersProjects
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

	if userExists(users, user, false) {
		resp := make(map[string]string)
		resp["msg"] = "User already exists"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	} else {
		users[user.Id] = hashAndSalt([]byte(user.Password))
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

// TODO Send projects
func login(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	user = getBodyUser(req)
	users := mapUserJson()
	if userExists(users, user, true) {
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
	http.HandleFunc("/createProject", createProject)
	http.HandleFunc("/getProject", getProject)
	err := http.ListenAndServe("localhost:443", nil)
	check(err)
}
