package model

import (
	"golang.org/x/crypto/bcrypt"

	validation "github.com/go-ozzo/ozzo-validation"
)

type User struct {
	ID                int    `json:"id,omitempty"`
	Login             string `json:"login,omitempty"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
	Role              int    `json:"role,omitempty"`
	Role_Name         string `json:"rolename"`
	FirstName         string `json:"name"`
	Lastname          string `json:"lastname"`
	Middlename        string `json:"middlename"`
	Module 			  int `json:"module"`
}

type Aeg struct {
	ID         int    `json:"id"`
	Firstname  string `json:"firstname"`
	Lastname   string `json:"lastname"`
	Middlename string `json:"middlename"`
	Role       string `json:"role"`
	Competence string `json:"competence"`
	Year       string `json:"year"`
}

type Competence struct {
	Id   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name"`
}
type Players struct {
	ID         int    `json:"id,omitempty"`
	Firstname  string `json:"firstname,omitempty"`
	Lastname   string `json:"lastname,omitempty"`
	Middlename string `json:"middlename,omitempty"`
	Gender     string `json:"gender,omitempty"`
	Dob        string `json:"dob,omitempty"`
	Inclusion  string `json:"inclusion,omitempty"`
	Year       string `json:"year,omitempty"`
	Leader     string `json:"leader,omitempty"`
	Competence string `json:"competence,omitempty"`
	Workplace  string `json:"workplace"`
}
type Tasks struct {
	ID         int    `json:"id,omitempty"`
	Jobnumber  string `json:"jobnumber"`
	Aspect     string `json:"aspect"`
	Maxpoint   int    `json:"maxpoint"`
	Module     int    `json:"module"`
	Competence string `json:"competence"`
	Year       string `json:"year"`
}
type TasksModel struct {
	Firstname  string `json:"firstname,omitempty"`
	Lastname   string `json:"lastname,omitempty"`
	Middlename string `json:"middlename,omitempty"`
	Module     int    `json:"module,omitempty"`
	Jobnumber  string `json:"jobnumber"`
	Aspect     string `json:"aspect"`
	Maxpoint   int    `json:"maxpoint"`
}
type PlayersModel struct {
	Firstname   string `json:"firstname,omitempty"`
	Lastname    string `json:"lastname,omitempty"`
	Middlename  string `json:"middlename,omitempty"`
	Module      int    `json:"module,omitempty"`
	FirstNameP  string `json:"firstnamep"`
	LastnameP   string `json:"lastnamep"`
	MiddlenameP string `json:"middlenamep"`
}
type Scoresheet struct {
	FirstnameP  string  `json:"firstnamep"`
	LastnameP   string  `json:"lastnamep"`
	MiddlenameP string  `json:"middlenamep"`
	FirstName   string  `json:"firstname"`
	Lastname    string  `json:"lastname"`
	Middlename  string  `json:"middlename"`
	Task        string  `json:"task"`
	Workplace   string  `json:"workplace"`
	Setpoint    float64 `json:"setpoint"`
}
type ListWinners struct {
	FirstName  string  `json:"firstname"`
	Lastname   string  `json:"lastname"`
	Middlename string  `json:"middlename"`
	Totalscore float64 `json:"totalscore"`
	Place      int     `json:"place"`
}

func (u *User) Validate() error {

	return validation.ValidateStruct(
		u,
		validation.Field(&u.Login, validation.Required, validation.Length(6, 45)),
		validation.Field(&u.Password, validation.By(requredif(u.EncryptedPassword == "")), validation.Length(6, 45)))
}
func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := encryptString(u.Password)
		if err != nil {
			return err
		}
		u.EncryptedPassword = enc
	}

	return nil
}

func (u *User) Sanitize() {
	u.Password = ""
	u.Role = 0
	u.ID = 0
	u.Login = ""
}

func (c *Competence) SanitizeC() {
	c.Id = 0
	c.Type = ""
}
func (ts *Tasks) SanitizeTasks() {
	ts.ID = 0
}
func (pl *PlayersModel) SanitizePlayers() {
	pl.Firstname = ""
	pl.Lastname = ""
	pl.Middlename = ""
	pl.Module = 0
}
func (pl *Players) SanitizePlayersW() {
	pl.Firstname = ""
	pl.Lastname = ""
	pl.Middlename = ""
	pl.Year = ""
	pl.Dob = ""
}
func (tsM *TasksModel) SanitizeTasksM() {
	tsM.Firstname = ""
	tsM.Lastname = ""
	tsM.Middlename = ""
	tsM.Module = 0
}
func encryptString(s string) (string, error) {

	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", nil
	}

	return string(b), nil
}

//сравнивает пароли который передал клиент и пароль который находиться в бд
func (u *User) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
