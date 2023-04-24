package main

/*
-------------------------------------------------------------------------------
							API del servidor
-------------------------------------------------------------------------------
*/

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	u "tasking_index_server/util"
)

// Assigns body parameters to a user
func GetBodyUser(req *http.Request) u.User {
	var user u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	error := json.Unmarshal(body, &user)
	u.Check(error)
	return user
}

func getBodyUserProject(req *http.Request) u.BodyUserProject {
	var project u.BodyUserProject
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	error := json.Unmarshal(body, &project)
	u.Check(error)
	return project
}

func getProjectId() map[string]int {
	data, fileErr := os.ReadFile("../data/projectId.json")
	u.Check(fileErr)
	projectIdJson := make(map[string]int)
	error := json.Unmarshal(data, &projectIdJson)
	u.Check(error)
	return projectIdJson
}

// Checks if user has TOTP correctly configured and enables 2FA
func check2FA(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var message string
	var user u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal(body, &user)
	users := u.StructUsersJson()
	if u.UserExists(users, user, true) {
		w.WriteHeader(200)
		savedUser := u.ObtainUser(user, users)
		if savedUser.DoubleAuthKey != "" {
			if u.CompareTOTPCode(savedUser.DoubleAuthKey, user.DoubleAuthCode) {
				savedUser.DoubleAuthActivated = true
				message = "2FA correctly set in your account"
			} else {
				message = "2FA temporal number doesn't match the one generated by the server"
			}
			users.Users[u.FindUser(users, savedUser)] = savedUser
			usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
			u.Check(JsonErr)
			erro := os.WriteFile("../data/users.json", usersJSON, 0666)
			u.Check(erro)
		} else {
			message = "2FA activation first step missing"
		}
	} else {
		w.WriteHeader(409)
		message = "User not found or incorrect password, could not activate 2FA(second step)"
	}
	resp := make(map[string]string)
	resp["msg"] = message
	jsonResp, jsonErr := json.Marshal(resp)
	u.Check(jsonErr)
	w.Write(jsonResp)
}

// Tries to enable 2FA on user, checkTOTP must be called later to definetely activate 2FA
func enable2FA(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal(body, &user)
	users := u.StructUsersJson()
	if u.UserExists(users, user, true) {
		w.WriteHeader(200)
		key := u.GenerateKey(16)
		totp := u.GenerateTOTP(key)
		savedUser := u.ObtainUser(user, users)
		savedUser.DoubleAuthKey = key
		users.Users[u.FindUser(users, savedUser)] = savedUser
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		u.Check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		u.Check(erro)
		url := totp.ProvisioningUri("tasking", "user")
		resp := make(map[string]string)
		resp["msg"] = "2FA first step completed, please check that your app generated code matches the server one"
		resp["url"] = url
		jsonResp, jsonErr := json.Marshal(resp)
		u.Check(jsonErr)
		w.Write(jsonResp)
	} else {
		w.WriteHeader(409)
		resp := make(map[string]string)
		resp["msg"] = "User not found or incorrect password, could not activate 2FA(first step)"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.Write(jsonResp)
	}
}

func disable2FA(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user u.User
	var message string
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal(body, &user)
	users := u.StructUsersJson()
	if u.UserExists(users, user, true) {
		w.WriteHeader(200)
		savedUser := u.ObtainUser(user, users)
		savedUser.DoubleAuthKey = ""
		savedUser.DoubleAuthActivated = false
		users.Users[u.FindUser(users, savedUser)] = savedUser
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		u.Check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		u.Check(erro)
		message = "2FA correctly disabled"
	} else {
		w.WriteHeader(409)
		message = "User not found or incorrect password, could not disable 2FA"
	}
	resp := make(map[string]string)
	resp["msg"] = message
	jsonResp, jsonErr := json.Marshal(resp)
	u.Check(jsonErr)
	w.Write(jsonResp)
}

// TODO añadir usuario a UsersProjects --> Está hecho CREO
// Preguntar a Angel si key a nil
func register(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user u.User
	//Read body and save in User Struct
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal(body, &user)

	// Read users.json and map
	data, fileErr := os.ReadFile("../data/users.json")
	u.Check(fileErr)
	var users u.Users
	json.Unmarshal(data, &users)

	if u.UserExists(users, user, false) {
		resp := make(map[string]string)
		resp["msg"] = "El usuario ya existe en la BD"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	} else {
		user.Password = u.HashAndSalt([]byte(user.Password))
		users.Users = append(users.Users, user)
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		u.Check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		u.Check(erro)
		resp := make(map[string]string)
		resp["msg"] = "Registro completado con éxito"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	}
}

func deleteProject(w http.ResponseWriter, req *http.Request) {
	users := u.StructUsersJson()
	var bodyUserProject u.BodyUserProject
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUserProject)
	if u.UserExists(users, bodyUserProject.User, true) {
		//Encontrar Id en users
		userIndex := u.FindUser(users, bodyUserProject.User)
		user := users.Users[userIndex]
		//Encontrar posicion del proyecto en users y eliminar del array
		projectIndex := u.FindProject(user, bodyUserProject.Project.Id)
		if projectIndex == -1 {
			resp := make(map[string]string)
			resp["msg"] = "El proyecto no pertenece a este usuario"
			jsonResp, respErr := json.Marshal(resp)
			u.Check(respErr)
			w.WriteHeader(409)
			w.Write(jsonResp)
			return
		}
		user.Projects = u.DisAppend(user.Projects, projectIndex)
		users.Users[userIndex] = user
		usersJson, usersErr := json.MarshalIndent(users, "", "  ")
		u.Check(usersErr)
		err := os.WriteFile("../data/users.json", usersJson, 0666)
		u.Check(err)
		//Eliminamos la carpeta
		filerr := os.RemoveAll("../projects/" + strconv.Itoa(bodyUserProject.Project.Id) + "/")
		u.Check(filerr)
		resp := make(map[string]string)
		resp["msg"] = "Proyecto borrado satisfactoriamente"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "El usuario o contraseña son incorrectos"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

// TODO --> Debe devolver los proyectos (como en login) y comprobar que el usuario tiene ese proyecto
func updateProject(w http.ResponseWriter, req *http.Request) {

	// Read users.json and map
	data, fileErr := os.ReadFile("../data/users.json")
	u.Check(fileErr)
	var users u.Users
	json.Unmarshal(data, &users)

	//Read files from multipart request
	//Se puede aumentar 10(KB)-->20(MB)-->30(GB)
	err := req.ParseMultipartForm(32 << 20) // maxMemory 32MB
	u.Check(err)
	body := req.FormValue("bodyJson")
	var bodyJson u.BodyUserProject
	json.Unmarshal([]byte(body), &bodyJson)
	if u.UserExists(users, bodyJson.User, true) {
		file, handler, err := req.FormFile("project")
		u.Check(err)
		filename := handler.Filename
		tmpfile, err := os.Create("../projects/" + strconv.Itoa(bodyJson.Project.Id) + "/" + filename)
		defer tmpfile.Close()
		u.Check(err)
		_, err = io.Copy(tmpfile, file)
		u.Check(err)
		resp := make(map[string]string)
		resp["msg"] = "Proyecto modificado satisfactoriamente"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "El usuario o contraseña son incorrectos"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

func createProject(w http.ResponseWriter, req *http.Request) {
	var body u.BodyUserProject
	body = getBodyUserProject(req)
	var bodyUser u.User
	bodyUser = body.User
	users := u.StructUsersJson()
	if u.UserExists(users, bodyUser, true) {
		projectId := getProjectId()
		id := projectId["projectId"]
		projectId["projectId"] = projectId["projectId"] + 1
		projectIdJson, projectIdErr := json.MarshalIndent(projectId, "", "  ")
		u.Check(projectIdErr)
		err := os.WriteFile("../data/projectId.json", projectIdJson, 0666)
		u.Check(err)
		resp := make(map[string]int)
		resp["id"] = id
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		stringId := strconv.Itoa(id)
		mkdirErr := os.MkdirAll("../projects/"+stringId, 0755)
		u.Check(mkdirErr)
		userIndex := u.FindUser(users, bodyUser)
		users.Users[userIndex].Projects = append(users.Users[userIndex].Projects, id)
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		u.Check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		u.Check(erro)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "El usuario o contraseña son incorrectos"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

func getKeys(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	if u.UserExists(users, bodyUser, true) {
		var keys u.Keys
		keys.Kpriv = users.Users[u.FindUser(users, bodyUser)].Keys.Kpriv
		keys.Kpub = users.Users[u.FindUser(users, bodyUser)].Keys.Kpub
		jsonResp, jsonErr := json.Marshal(keys)
		u.Check(jsonErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "El usuario o contraseña son incorrectos"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

func getProjects(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	savedUser := u.ObtainUser(bodyUser, users)
	if !u.TOTPactivated(savedUser) || u.CompareTOTPCode(savedUser.DoubleAuthKey, bodyUser.DoubleAuthCode) {
		//Escribimos el zip sobre la respuesta
		writer := zip.NewWriter(w)
		defer writer.Close()
		filenames := u.GetFilenames(bodyUser, users)
		//Recorremos todos los proyectos guardados
		for _, filename := range filenames {
			//Hacemos una copia de los ficheros en el zip
			file, err := os.Open(filename)
			u.Check(err)
			defer file.Close()
			newFilename := strings.Split(filename, "/")
			fileToZip, err := writer.Create(newFilename[3])
			u.Check(err)
			_, err = io.Copy(fileToZip, file)
			u.Check(err)
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "data"))
		w.WriteHeader(200)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "Código de 2FA incorrecto"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

func login(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var bodyUser u.User
		body, reqErr := io.ReadAll(req.Body)
		u.Check(reqErr)
		json.Unmarshal([]byte(body), &bodyUser)
		users := u.StructUsersJson()
		if !u.UserExists(users, bodyUser, true) {
			resp := make(map[string]string)
			resp["msg"] = "El usuario o contraseña son incorrectos"
			jsonResp, respErr := json.Marshal(resp)
			u.Check(respErr)
			w.WriteHeader(409)
			w.Write(jsonResp)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func main() {
	server := "localhost:8080"
	fmt.Println("Servidor a la espera de peticiones en " + server)
	mux := http.NewServeMux()
	loginHandler := http.HandlerFunc(getProjects)
	http.HandleFunc("/register", register)
	mux.Handle("/login", login(loginHandler))
	http.HandleFunc("/createProject", createProject)
	http.HandleFunc("/updateProject", updateProject)
	http.HandleFunc("/deleteProject", deleteProject)
	http.HandleFunc("/enable2FA", enable2FA)
	http.HandleFunc("/check2FA", check2FA)
	http.HandleFunc("/disable2FA", disable2FA)
	http.HandleFunc("/getKeys", getKeys)
	//http.HandleFunc("/getProject", getProject)

	err := http.ListenAndServe(server, mux)
	u.Check(err)
}
