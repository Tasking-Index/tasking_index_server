package main

/*
-------------------------------------------------------------------------------
							API del servidor
-------------------------------------------------------------------------------
*/

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
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

/*
func getProject(w http.ResponseWriter, req *http.Request) {
		body, reqErr := io.ReadAll(req.Body)
		u.Check(reqErr)
		users := u.StructUsersJson()
		var user User
		json.Unmarshal(body, &user)
		if u.UserExists(users, user, true) {
			//DAR INFORMACIÓN DEL PROYECTO
			projects := getProjectsJson()
			fmt.Println(projects)
		} else {
			resp := make(map[string]string)
			resp["msg"] = "Login needed to get projects"
			jsonResp, respErr := json.Marshal(resp)
			u.Check(respErr)
			w.WriteHeader(400)
			w.Write(jsonResp)
		}
}
*/

// Checks if user has TOTP correctly configured and activates 2FA if so
/*ÁNGEL ESTÁ COMENTADO PORQUE PEGA PETARDAZO
func checkTOTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user userTOTP
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal(body, &user)
	users := u.StructUsersJson()
	if u.UserExists(users, user, true) {
		w.WriteHeader(200)
		savedUser := u.ObtainUser(user, users)
		key := savedUser.DoubleAuthKey
		totp := u.GenerateTOTP(key)
		totpCode := totp.Now()
		if {

		}
		users.Users[u.FindUser(users, savedUser)] = savedUser
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		u.Check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		u.Check(erro)
		resp := make(map[string]string)
		resp["msg"] = "2FA correctly set in your account"
		jsonUrl, jsonErr := json.Marshal(resp)
		u.Check(jsonErr)
		w.Write(jsonUrl)
	} else {
		w.WriteHeader(409)
		resp := make(map[string]string)
		resp["msg"] = "User not found or incorrect password, could not activate 2FA(second step)"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.Write(jsonResp)
	}
}
*/

// Tries to enable 2FA on user, checkTOTP must be called later to definetely activate 2FA
func activateTOTP(w http.ResponseWriter, req *http.Request) {
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
		resp["url"] = url
		jsonUrl, jsonErr := json.Marshal(resp)
		u.Check(jsonErr)
		w.Write(jsonUrl)
	} else {
		w.WriteHeader(409)
		resp := make(map[string]string)
		resp["msg"] = "User not found or incorrect password, could not activate 2FA(first step)"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.Write(jsonResp)
	}
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

func login(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	if u.UserExists(users, bodyUser, true) {
		//Obtenemos el nombre de los zips y lo enviamos en la respuesta
		filenames := u.GetFilenames(bodyUser, users)
		for _, filename := range filenames {
			file, err := os.Open(filename)
			u.Check(err)
			_, err = io.Copy(w, file)
			u.Check(err)
		}
		w.WriteHeader(200)
	} else {
		resp := make(map[string]string)
		resp["msg"] = "El usuario o contraseña son incorrectos"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
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

// TODO --> Debe devolver los proyectos (como en login)
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
	body = req.FormValue("files")
	fileKeysJson := make(map[string][]string)
	json.Unmarshal([]byte(body), &fileKeysJson)
	fileKeys := fileKeysJson["filekeys"]
	if u.UserExists(users, bodyJson.User, true) {
		for _, fileKey := range fileKeys {
			file, handler, err := req.FormFile(fileKey)
			u.Check(err)
			filename := handler.Filename
			tmpfile, err := os.Create("../projects/" + strconv.Itoa(bodyJson.Project.Id) + "/" + filename)
			defer tmpfile.Close()
			u.Check(err)
			_, err = io.Copy(tmpfile, file)
			u.Check(err)
		}
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

func main() {
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/createProject", createProject)
	http.HandleFunc("/updateProject", updateProject)
	http.HandleFunc("/deleteProject", deleteProject)
	//http.HandleFunc("/getProject", getProject)
	err := http.ListenAndServe("192.168.1.103:443", nil)
	u.Check(err)
}

// Código para enviar archivos
func handleFile(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("archivo.txt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/plain")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
