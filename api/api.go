package main

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	Users []User `json:"users"`
}

type BodyUserProject struct {
	User    User    `json:"bodyUser"`
	Project Project `json:"bodyProject"`
}

type User struct {
	Id                  string `json:"user"`
	Password            string `json:"pass"`
	Kpriv               string `json:"kPriv"`
	Kpub                string `json:"kPub"`
	DoubleAuthKey       string `json:"doubleAuthKey"`
	DoubleAuthActivated bool   `json:"doubleAuthActivated"`
	Projects            []int  `json:"projects"`
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

func userExists(users Users, user User, needPassword bool) bool {
	for _, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			if needPassword {
				pswd := user.Password
				ok := comparePasswords(userSaved.Password, []byte(pswd))
				return ok
			}
			return true
		}
	}
	return false
}

func findUser(users Users, user User) int {
	var index int
	for i, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			index = i
			break
		}
	}
	return index
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
	return string(hash)
}

func comparePasswords(hashedPwd string, bytePwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePwd)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Structs users.json
func StructUsersJson() Users {
	data, fileErr := os.ReadFile("../data/users.json")
	check(fileErr)
	var users Users
	error := json.Unmarshal(data, &users)
	check(error)
	return users
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

func getBodyUserProject(req *http.Request) BodyUserProject {
	var project BodyUserProject
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	error := json.Unmarshal(body, &project)
	check(error)
	return project
}

func getProjectId() map[string]int {
	data, fileErr := os.ReadFile("../data/projectId.json")
	check(fileErr)
	projectIdJson := make(map[string]int)
	error := json.Unmarshal(data, &projectIdJson)
	check(error)
	return projectIdJson
}

/*
func getProject(w http.ResponseWriter, req *http.Request) {
	print("HOLA")
	body, reqErr := io.ReadAll(req.Body)
	check(reqErr)
	users := StructUsersJson()
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
*/

// TODO añadir usuario a UsersProjects
// Preguntar a Angel si key a nil
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
	var users Users
	json.Unmarshal(data, &users)

	if userExists(users, user, false) {
		resp := make(map[string]string)
		resp["msg"] = "User already exists"
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	} else {
		user.Password = hashAndSalt([]byte(user.Password))
		users.Users = append(users.Users, user)
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
	//user.Password = hashAndSalt([]byte(user.Password))
	users := StructUsersJson()
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

func updateProject(w http.ResponseWriter, req *http.Request) {

}

func createProject(w http.ResponseWriter, req *http.Request) {
	var body BodyUserProject
	body = getBodyUserProject(req)
	var bodyUser User
	bodyUser = body.User
	users := StructUsersJson()
	if userExists(users, bodyUser, true) {
		projectId := getProjectId()
		id := projectId["projectId"]
		projectId["projectId"] = projectId["projectId"] + 1
		projectIdJson, projectIdErr := json.MarshalIndent(projectId, "", "  ")
		check(projectIdErr)
		err := os.WriteFile("../data/projectId.json", projectIdJson, 0666)
		check(err)
		resp := make(map[string]int)
		resp["id"] = id
		jsonResp, respErr := json.Marshal(resp)
		check(respErr)
		stringId := strconv.Itoa(id)
		mkdirErr := os.MkdirAll("../projects/"+stringId, 0755)
		check(mkdirErr)
		userIndex := findUser(users, bodyUser)
		users.Users[userIndex].Projects = append(users.Users[userIndex].Projects, id)
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		check(erro)
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
	http.HandleFunc("/updateProject", updateProject)
	//http.HandleFunc("/getProject", getProject)
	err := http.ListenAndServe("localhost:443", nil)
	check(err)
}
