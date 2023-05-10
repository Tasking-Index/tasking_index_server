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
	Id                  string  `json:"user"`
	Password            string  `json:"pass"`
	Keys                Keys    `json:"keys"`
	DoubleAuthKey       string  `json:"doubleAuthKey"`
	DoubleAuthActivated bool    `json:"doubleAuthActivated"`
	DoubleAuthCode      string  `json:"totpCode"`
	Projects            []int   `json:"projects"`
	Friends             Friends `json:"friends"`
}

type TokenUser struct {
	Token               string  `json:"token"`
	Password            string  `json:"pass"`
	Keys                Keys    `json:"keys"`
	DoubleAuthKey       string  `json:"doubleAuthKey"`
	DoubleAuthActivated bool    `json:"doubleAuthActivated"`
	DoubleAuthCode      string  `json:"totpCode"`
	Projects            []int   `json:"projects"`
	Friends             Friends `json:"friends"`
}

/*
Proyecto con sus datos
*/
type Project struct {
	Id          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Files       []string   `json:"files"`
	Additional  Additional `json:"additional"`
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

type Keys struct {
	Kpriv  string `json:"kPriv"`
	Kpub   string `json:"kPub"`
	IVpriv string `json:"ivPriv"`
}

type Friends struct {
	Available []string `json:"available"`
	Requested []string `json:"requested"`
	Pending   []string `json:"pending"`
}
