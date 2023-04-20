package util

/*
--------------------------------------------------------------------------------
				Estructuras necesarias para usar la API
--------------------------------------------------------------------------------
*/

/*
Estructura auxiliar para recuperar un listado de usuarios de la API
*/
type Users struct {
	Users []User `json:"users"`
}

/*
Estructura auxiliar para las peticiones de la API:

	"bodyUser":{
		"user": "user1",
		"pass": "chinin1"
	},

	"bodyProject":{
		"title:": "Proyecto"
	}
*/
type BodyUserProject struct {
	User    User    `json:"bodyUser"`
	Project Project `json:"bodyProject"`
}

/*
Usuario con sus datos
*/
type User struct {
	Id                  string `json:"user"`
	Password            string `json:"pass"`
	Kpriv               string `json:"kPriv"`
	Kpub                string `json:"kPub"`
	DoubleAuthKey       string `json:"doubleAuthKey"`
	DoubleAuthActivated bool   `json:"doubleAuthActivated"`
	DoubleAuthCode 		string `json:"totpCode"`
	Projects            []int  `json:"projects"`
}

/*
Proyecto con sus datos
*/
type Project struct {
	Id          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
}

/*
Campos adicionales de un proyecto
*/
type Additional struct {
	fecha_inicio  string
	fecha_fin     string
	propietario   string
	colaboradores []string
	tareas        []Tarea
	logs          []string
}

/*
Informacion de una tarea
*/
type Tarea struct {
	nombre string
	estado bool
}
