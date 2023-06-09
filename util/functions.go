package util

/*
--------------------------------------------------------------------------------
		Funciones necesarias para realizar comprobaciones y almacenar en
						los JSON la informacion
--------------------------------------------------------------------------------
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt"
	"github.com/xlzd/gotp"
	"golang.org/x/crypto/bcrypt"
)

// taskingindexserver en SHA256
var HMACSECRET = []byte("61394ba0cc7da0efa813264559fbe061df7e58b99e8d2e5eb9a831306ed64fe2")

/*
Controla si hay un error

	Parametros	(error)
	Devuelve	Error si lo hay
				Continua su ejecucion con normalidad
*/
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func FindProject(user User, id int) int {
	pos := 0
	for _, project := range user.Projects {
		if project == id {
			return pos
		}
		pos = pos + 1
	}
	return -1
}

func DisAppendInt(slice []int, index int) []int {
	return append(slice[:index], slice[index+1:]...)
}

func DisAppendString(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

/*
Devuelve el nombre de los ficheros de proyectos
*/
func GetFilenames(user User, users Users) []string {
	projectIds := GetProjects(user, users)
	var filenames []string
	for _, projectId := range projectIds {
		//Puede que haya que cambiarlo
		name := "../projects/" + strconv.Itoa(projectId) + "/" + "project" + strconv.Itoa(projectId) + ".zip"
		filenames = append(filenames, name)
	}
	return filenames
}

/*
Obtiene los índices de los proyectos de un usuario
[Debe haberse comprobado anteriormente que existe]
*/
func GetProjects(user User, users Users) []int {
	var projectIds []int
	for _, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			projectIds = userSaved.Projects
		}
	}
	return projectIds
}

/*
Generates a temporal TOTP object
*/
func GenerateTOTP(secretKey string) *gotp.TOTP {
	totp := gotp.NewDefaultTOTP(secretKey)
	return totp
}

/*
Generates a key, should be saved in user account
*/
func GenerateKey(length int) string {
	secretKey := gotp.RandomSecret(length)
	return secretKey
}

func CompareTOTPCode(secret string, totpCode string) bool {
	same := false
	totp := GenerateTOTP(secret)
	totpServerCode := totp.Now()
	if totpServerCode == totpCode {
		same = true
	}
	return same
}

func TOTPactivated(user User) bool { return user.DoubleAuthActivated }

func GetUserByToken(tokenUser TokenUser) User {
	var user User
	token, _ := jwt.Parse(tokenUser.Token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return HMACSECRET, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user.Id = claims["user"].(string)
		user.Password = claims["pass"].(string)
		return user
	}
	return user
}

/*
Valida que un usuario existe junto con su password:

	Parametros	(Lista Users, User, Es necesario el password?)
	Devuelve	Verdadero	si existe o...
				Falso 		en caso contrario
*/
func UserExists(users Users, user User, needPassword bool) bool {
	for _, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			if needPassword {
				ok := ComparePasswords(userSaved.Password, []byte(user.Password))
				return ok
			}
			return true
		}
	}
	return false
}

func HasProject(user User, id int) (bool, int) {
	for i, project := range user.Projects {
		if id == project {
			return true, i
		}
	}
	return false, -1
}

func ObtainUser(user User, users Users) User {
	var searchedUser User
	for _, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			if ComparePasswords(userSaved.Password, []byte(user.Password)) {
				searchedUser = userSaved
			}
		}
	}
	return searchedUser
}

func ObtainPublicKey(user User, users Users) map[string]string {
	resp := make(map[string]string)
	for _, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			resp["publicKey"] = userSaved.Keys.Kpub
			return resp
		}
	}
	return resp
}

/*
Obtiene la posicion de un Usuario en la lista que previamente sabemos que existe, sino obtendremos error

	Parametros	(Lista Users, User)
	Devuelve	Posicion del User en la lista
*/
func FindUser(users Users, user User) int {
	var index int = -1
	for i, userSaved := range users.Users {
		if userSaved.Id == user.Id {
			index = i
			break
		}
	}
	return index
}

/*
Comprueba si un slice contiene un determinado valor

	Parametros	(slice, string)
	Devuelve	Si el string existe o no en el slice y su posición
*/
func ContainsString(s []string, str string) (bool, int) {
	for i, v := range s {
		if v == str {
			return true, i
		}
	}
	return false, -1
}

func ContainsInt(s []int, str int) (bool, int) {
	for i, v := range s {
		if v == str {
			return true, i
		}
	}
	return false, -1
}

/*
Hashea un password y devuelve lo devuelve compuesto como una string

	Parametros	(cadena)
	Devuelve	cadena hasheada
*/
func HashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

/*
Valida que un hash y un password son lo mismo

	Parametros	(hash, cadena)
	Devuelve	Verdadero	si son lo mismo o...
				Falso		en caso contrario
*/
func ComparePasswords(hashedPwd string, bytePwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePwd)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

/*
Almacena la lista de usuarios del archivo ""../data/users.json" en una estructura del tipo Users

	Devuelve	Estructura Users
*/
func StructUsersJson() Users {
	data, fileErr := os.ReadFile("../data/users.json")
	Check(fileErr)
	var users Users
	error := json.Unmarshal(data, &users)
	Check(error)
	return users
}
