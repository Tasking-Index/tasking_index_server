package main

/*
-------------------------------------------------------------------------------
							API del servidor
-------------------------------------------------------------------------------
*/

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
}

func disable2FA(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user u.User
	var message string
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal(body, &user)
	users := u.StructUsersJson()
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
	user.Projects = u.DisAppendInt(user.Projects, projectIndex)
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
}

// TODO --> ¿Debe devolver los proyectos (como en login) [Lo puede actualizar Chinin en local]? y comprobar que el usuario tiene ese proyecto
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
}

func createProject(w http.ResponseWriter, req *http.Request) {
	var body u.BodyUserProject
	body = getBodyUserProject(req)
	var bodyUser u.User
	bodyUser = body.User
	users := u.StructUsersJson()
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
}

func addColaborator(w http.ResponseWriter, req *http.Request) {
	var body u.BodyUserProject
	var user u.User
	var project u.Project
	var user2 u.User
	body = getBodyUserProject(req)
	user = body.User
	project = body.Project
	users := u.StructUsersJson()
	user2.Id = user.Friends.Available[0]
	existsFriend, _ := u.Contains(users.Users[u.FindUser(users, user)].Friends.Available, user2.Id)
	existsProject := u.HasProject(users.Users[u.FindUser(users, user2)], project.Id)

	print("Hi")

	resp := make(map[string]string)
	if existsFriend && !existsProject {
		users.Users[u.FindUser(users, user2)].Projects = append(users.Users[u.FindUser(users, user2)].Projects, project.Id)
		usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
		u.Check(JsonErr)
		erro := os.WriteFile("../data/users.json", usersJSON, 0666)
		u.Check(erro)
		resp["msg"] = "Colaborador añadido correctamente"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(200)
		w.Write(jsonResp)
	} else {
		resp["msg"] = "ERROR: el colaborador no es amigo"
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	}
}

func getUsers(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()

	resp := make(map[string][]string)
	var ids []string
	for id, user := range users.Users {
		if bodyUser.Id != user.Id {
			ids = append(ids, users.Users[id].Id)
		}
	}
	resp["users"] = ids
	jsonResp, respErr := json.Marshal(resp)
	u.Check(respErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func getFriends(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	var friends u.Friends
	friends = users.Users[u.FindUser(users, bodyUser)].Friends
	jsonResp, respErr := json.Marshal(friends)
	u.Check(respErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func deleteFriends(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	var auxUser u.User
	for _, id := range bodyUser.Friends.Available {
		auxUser.Id = id
		exists, sourcePos := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Available, id)
		_, targetPos := u.Contains(users.Users[u.FindUser(users, auxUser)].Friends.Available, bodyUser.Id)
		if exists {
			users.Users[u.FindUser(users, bodyUser)].Friends.Available = u.DisAppendString(users.Users[u.FindUser(users, bodyUser)].Friends.Available, sourcePos)
			users.Users[u.FindUser(users, auxUser)].Friends.Available = u.DisAppendString(users.Users[u.FindUser(users, auxUser)].Friends.Available, targetPos)
		}
	}
	usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
	u.Check(JsonErr)
	erro := os.WriteFile("../data/users.json", usersJSON, 0666)
	u.Check(erro)
	resp := make(map[string]string)
	resp["msg"] = "Amigos eliminados satisfactoriamente"
	jsonResp, respErr := json.Marshal(resp)
	u.Check(respErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func acceptFriends(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	var auxUser u.User
	for _, id := range bodyUser.Friends.Available {
		exists, _ := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Available, id)
		if !exists {
			auxUser.Id = id
			_, pendingPos := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Pending, id)
			_, requestedPos := u.Contains(users.Users[u.FindUser(users, auxUser)].Friends.Requested, bodyUser.Id)
			users.Users[u.FindUser(users, bodyUser)].Friends.Available = append(users.Users[u.FindUser(users, bodyUser)].Friends.Available, id)
			users.Users[u.FindUser(users, auxUser)].Friends.Available = append(users.Users[u.FindUser(users, auxUser)].Friends.Available, bodyUser.Id)
			users.Users[u.FindUser(users, bodyUser)].Friends.Pending = u.DisAppendString(users.Users[u.FindUser(users, bodyUser)].Friends.Pending, pendingPos)
			users.Users[u.FindUser(users, auxUser)].Friends.Requested = u.DisAppendString(users.Users[u.FindUser(users, auxUser)].Friends.Requested, requestedPos)
		}
	}
	usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
	u.Check(JsonErr)
	erro := os.WriteFile("../data/users.json", usersJSON, 0666)
	u.Check(erro)
	resp := make(map[string]string)
	resp["msg"] = "Solicitudes de amistad aceptadas"
	jsonResp, respErr := json.Marshal(resp)
	u.Check(respErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func rejectFriends(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	var auxUser u.User
	for _, id := range bodyUser.Friends.Pending {
		exists, _ := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Pending, id)
		if exists {
			auxUser.Id = id
			_, pendingPos := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Pending, id)
			_, requestedPos := u.Contains(users.Users[u.FindUser(users, auxUser)].Friends.Requested, bodyUser.Id)
			users.Users[u.FindUser(users, bodyUser)].Friends.Pending = u.DisAppendString(users.Users[u.FindUser(users, bodyUser)].Friends.Pending, pendingPos)
			users.Users[u.FindUser(users, auxUser)].Friends.Requested = u.DisAppendString(users.Users[u.FindUser(users, auxUser)].Friends.Requested, requestedPos)
		}
	}
	usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
	u.Check(JsonErr)
	erro := os.WriteFile("../data/users.json", usersJSON, 0666)
	u.Check(erro)
	resp := make(map[string]string)
	resp["msg"] = "Solicitudes de amistad aceptadas"
	jsonResp, respErr := json.Marshal(resp)
	u.Check(respErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

// Añade usuarios de un array al campo requested y añade a pending de dichos usuarios al usuario que realiza la petición
func friendRequests(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	var auxUser u.User
	for _, id := range bodyUser.Friends.Requested {
		existsRequested, _ := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Requested, id)
		existsAvailable, _ := u.Contains(users.Users[u.FindUser(users, bodyUser)].Friends.Available, id)
		if !existsRequested && !existsAvailable {
			auxUser.Id = id
			users.Users[u.FindUser(users, bodyUser)].Friends.Requested = append(users.Users[u.FindUser(users, bodyUser)].Friends.Requested, id)
			users.Users[u.FindUser(users, auxUser)].Friends.Pending = append(users.Users[u.FindUser(users, auxUser)].Friends.Pending, bodyUser.Id)
		}
	}
	usersJSON, JsonErr := json.MarshalIndent(users, "", "  ")
	u.Check(JsonErr)
	erro := os.WriteFile("../data/users.json", usersJSON, 0666)
	u.Check(erro)
	resp := make(map[string]string)
	resp["msg"] = "Solicitudes de amistad enviadas satisfactoriamente"
	jsonResp, respErr := json.Marshal(resp)
	u.Check(respErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func getKeys(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
	var keys u.Keys
	keys.Kpriv = users.Users[u.FindUser(users, bodyUser)].Keys.Kpriv
	keys.Kpub = users.Users[u.FindUser(users, bodyUser)].Keys.Kpub
	keys.IVpriv = users.Users[u.FindUser(users, bodyUser)].Keys.IVpriv
	jsonResp, jsonErr := json.Marshal(keys)
	u.Check(jsonErr)
	w.WriteHeader(200)
	w.Write(jsonResp)
}

func getProjects(w http.ResponseWriter, req *http.Request) {
	var bodyUser u.User
	body, reqErr := io.ReadAll(req.Body)
	u.Check(reqErr)
	json.Unmarshal([]byte(body), &bodyUser)
	users := u.StructUsersJson()
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
}

func login(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var bodyUser u.User
		var message string
		body, reqErr := io.ReadAll(req.Body)
		u.Check(reqErr)
		reqCopy := ioutil.NopCloser(bytes.NewBuffer(body))
		req.Body = reqCopy
		json.Unmarshal([]byte(body), &bodyUser)
		users := u.StructUsersJson()
		savedUser := u.ObtainUser(bodyUser, users)
		if u.UserExists(users, bodyUser, true) {
			if !u.TOTPactivated(savedUser) || u.CompareTOTPCode(savedUser.DoubleAuthKey, bodyUser.DoubleAuthCode) {
				next.ServeHTTP(w, req)
				return
			} else {
				message = "2FA code does not match the server one"
			}
		} else {
			message = "User not found or incorrect password"
		}
		resp := make(map[string]string)
		resp["msg"] = message
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	})
}

func loginProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var bodyUserProject u.BodyUserProject
		var message string
		body, reqErr := io.ReadAll(req.Body)
		u.Check(reqErr)
		reqCopy := ioutil.NopCloser(bytes.NewBuffer(body))
		req.Body = reqCopy
		json.Unmarshal([]byte(body), &bodyUserProject)
		users := u.StructUsersJson()
		savedUser := u.ObtainUser(bodyUserProject.User, users)
		if u.UserExists(users, bodyUserProject.User, true) {
			if !u.TOTPactivated(savedUser) || u.CompareTOTPCode(savedUser.DoubleAuthKey, bodyUserProject.User.DoubleAuthCode) {
				next.ServeHTTP(w, req)
				return
			} else {
				message = "2FA code does not match the server one"
			}
		} else {
			message = "User not found or incorrect password"
		}
		resp := make(map[string]string)
		resp["msg"] = message
		jsonResp, respErr := json.Marshal(resp)
		u.Check(respErr)
		w.WriteHeader(409)
		w.Write(jsonResp)
	})
}

func main() {
	server := "192.168.68.101:8080"
	fmt.Println("Servidor a la espera de peticiones en " + server)
	mux := http.NewServeMux()
	registerHandler := http.HandlerFunc(register)
	mux.Handle("/register", registerHandler)
	getProjectsHandler := http.HandlerFunc(getProjects)
	mux.Handle("/login", login(getProjectsHandler))
	createProjectHandler := http.HandlerFunc(createProject)
	mux.Handle("/createProject", loginProject(createProjectHandler))
	//Update va a petar (recibe multipart, no bodyuserproject en loginProject)
	updateProjectHandler := http.HandlerFunc(updateProject)
	mux.Handle("/updateProject", loginProject(updateProjectHandler))
	deleteProjectHandler := http.HandlerFunc(deleteProject)
	mux.Handle("/deleteProject", loginProject(deleteProjectHandler))
	enable2FAHandler := http.HandlerFunc(enable2FA)
	mux.Handle("/enable2FA", login(enable2FAHandler))
	check2FAHandler := http.HandlerFunc(check2FA)
	mux.Handle("/check2FA", login(check2FAHandler))
	disable2FAHandler := http.HandlerFunc(disable2FA)
	mux.Handle("/disable2FA", login(disable2FAHandler))
	getKeysHandler := http.HandlerFunc(getKeys)
	mux.Handle("/getKeys", login(getKeysHandler))
	friendRequestsHandler := http.HandlerFunc(friendRequests)
	mux.Handle("/friendRequests", login(friendRequestsHandler))
	acceptFriendsHandler := http.HandlerFunc(acceptFriends)
	mux.Handle("/acceptFriends", login(acceptFriendsHandler))
	rejectFriendsHandler := http.HandlerFunc(rejectFriends)
	mux.Handle("/rejectFriends", login(rejectFriendsHandler))
	deleteFriendsHandler := http.HandlerFunc(deleteFriends)
	mux.Handle("/deleteFriends", login(deleteFriendsHandler))
	getFriendsHandler := http.HandlerFunc(getFriends)
	mux.Handle("/getFriends", login(getFriendsHandler))
	getUsersHandler := http.HandlerFunc(getUsers)
	mux.Handle("/getUsers", login(getUsersHandler))
	addColaboratorHandler := http.HandlerFunc(addColaborator)
	mux.Handle("/addColaborator", login(addColaboratorHandler))
	err := http.ListenAndServe(server, mux)
	u.Check(err)
}
