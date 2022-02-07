package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/Calavrat/http-rest-api/internal/app/model"
	"github.com/Calavrat/http-rest-api/internal/app/store"
	"github.com/sethvargo/go-password/password"
	"github.com/tealeg/xlsx"
	"golang.org/x/crypto/bcrypt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type UserRepository struct {
	store *Store
}

//Create
func (r UserRepository) Create(u *model.User) error {

	if err := u.Validate(); err != nil {
		return err
	}
	if err := u.BeforeCreate(); err != nil {
		return err
	}

	rows, err := r.store.db.Query("select max(id) FROM vkr.authorization")
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&u.ID)
		}
	}
	u.ID++

	if _, err = r.store.db.Exec(
		"INSERT INTO vkr.authorization (Id,Login,Password) VALUES(?,?,?)",
		u.ID,
		u.Login,
		u.EncryptedPassword,
	); err != nil {
		return err
	}

	return err
}

func (r *UserRepository) SelectComT(Type string) ([]model.Competence, error) {
	c := model.Competence{}
	comT := []model.Competence{}
	rows, err := r.store.db.Query("SELECT b.id, a.name, b.name FROM vkr.competence as b, vkr.type as a WHERE a.competence_id = b.id and a.name = ?", Type)
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&c.Id, &c.Type, &c.Name)
		if err != nil {
			continue
		}
		comT = append(comT, c)
	}

	return comT, err
}

func (r *UserRepository) CreateUsersG(me *model.Aeg, rc int) error {
	var (
		role_id int
		com_id  int
		aut_id  int
		id int
		con_id int
	)

rows, err := r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
		me.Competence)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}
	rows, err = r.store.db.Query("SELECT id FROM vkr.users WHERE  firstname = ? and  lastname = ? and middlename = ? and year = ? and competence_id = ?",
		me.Firstname, me.Lastname, me.Middlename, me.Year, com_id)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&id)
		}
	}
	if id != 0 {
		rows, err = r.store.db.Query("SELECT id FROM vkr.role WHERE name = ? ",
			me.Role)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&role_id)
			}
		}
		rows, err = r.store.db.Query("SELECT authorization_id FROM vkr.users WHERE firstname = ? and lastname = ? and middlename = ?",
			me.Firstname, me.Lastname, me.Middlename)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&aut_id)
			}
		}
		if _, err := r.store.db.Exec(
			"UPDATE vkr.users SET firstname = ?,  lastname = ?,  middlename= ?,  year = ?,   role_id = ?,  authorization_id= ?, competence_id = ?  WHERE id = ?",
			me.Firstname,
			me.Lastname,
			me.Middlename,
			me.Year,
			role_id,
			aut_id,
			com_id,
			id,
		); err != nil {
			return err
		}
	} else {
	rows, err = r.store.db.Query("SELECT MAX(id) FROM vkr.contest")
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&con_id)
			}
		}
		con_id++
	if _, err := r.store.db.Exec(
		"INSERT INTO vkr.contest (id,year,competence_id) VALUES (?,?,?)",
		con_id,
		me.Year,
		com_id,
	); err != nil {
		return err
	}
	rand.Seed(time.Now().UnixNano())
	chars := []rune("0123456789")
	length := 3
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	login := me.Firstname
	login += b.String()

	password, err := password.Generate(8, 4, 0, false, false)
	if err != nil {
		return err
	}

	var file *xlsx.File
	var cell *xlsx.Cell

	excelFileName := "MainUsersReg.xlsx"

	file, err = xlsx.OpenFile(excelFileName)
	if err != nil {
		return err
	}
	leader := me.Firstname + " " + me.Lastname + " " + me.Middlename
	for _, sheet := range file.Sheets {
		cell = AddCell(sheet, 0, 0)
			cell.Value = "ФИО"
			cell = AddCell(sheet, 0, 1)
			cell.Value = "Логин"
			cell = AddCell(sheet, 0, 2)
			cell.Value = "Пароль"
			cell = AddCell(sheet, 0, 3)
			cell.Value = "Компетенция"
			cell = AddCell(sheet, rc, 0)
			cell.Value = leader
			cell = AddCell(sheet, rc, 1)
			cell.Value = login
			cell = AddCell(sheet, rc, 2)
			cell.Value = password
			cell = AddCell(sheet, rc, 3)
			cell.Value = me.Competence

	}
	file.Save("MainUsersReg.xlsx")

	encpassword, err := encryptString(password)
	if err != nil {
		return err
	}

	rows, err := r.store.db.Query("SELECT MAX(id) FROM vkr.authorization")
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&aut_id)
		}
	}
	aut_id++

	if _, err := r.store.db.Exec(
		"INSERT INTO vkr.authorization (id, login, password) VALUES (?,?,?)",
		aut_id,
		login,
		encpassword,
	); err != nil {
		return err
	}

	rows, err = r.store.db.Query("SELECT MAX(id) FROM vkr.users")
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&me.ID)
		}
	}
	me.ID++

	rows, err = r.store.db.Query("SELECT id FROM vkr.role WHERE name = ? ",
		me.Role)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&role_id)
		}
	}

	rows, err = r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
		me.Competence)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}

	if _, err := r.store.db.Exec(
		"INSERT INTO vkr.users (id, firstname, lastname, middlename, year, role_id, authorization_id, competence_id) VALUES (?,?,?,?,?,?,?,?)",
		me.ID,
		me.Firstname,
		me.Lastname,
		me.Middlename,
		me.Year,
		role_id,
		aut_id,
		com_id,
	); err != nil {
		return err
	}
}
	return err
}

func (r *UserRepository) CreateUsers(me *model.Aeg, rc int) error {
	var (
		role_id int
		com_id  int
		aut_id  int
		id      int
	)
	rows, err := r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
		me.Competence)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}
	rows, err = r.store.db.Query("SELECT id FROM vkr.users WHERE  firstname = ? and  lastname = ? and middlename = ? and year = ? and competence_id = ?",
		me.Firstname, me.Lastname, me.Middlename, me.Year, com_id)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&id)
		}
	}
	if id != 0 {
		rows, err = r.store.db.Query("SELECT id FROM vkr.role WHERE name = ? ",
			me.Role)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&role_id)
			}
		}
		rows, err = r.store.db.Query("SELECT authorization_id FROM vkr.users WHERE firstname = ? and lastname = ? and middlename = ?",
			me.Firstname, me.Lastname, me.Middlename)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&aut_id)
			}
		}
		if _, err := r.store.db.Exec(
			"UPDATE vkr.users SET firstname = ?,  lastname = ?,  middlename= ?,  year = ?,   role_id = ?,  authorization_id= ?, competence_id = ?  WHERE id = ?",
			me.Firstname,
			me.Lastname,
			me.Middlename,
			me.Year,
			role_id,
			aut_id,
			com_id,
			id,
		); err != nil {
			return err
		}
	} else {
		rand.Seed(time.Now().UnixNano())
		chars := []rune("0123456789")
		length := 3
		var b strings.Builder
		for i := 0; i < length; i++ {
			b.WriteRune(chars[rand.Intn(len(chars))])
		}
		login := me.Firstname
		login += b.String()

		password, err := password.Generate(8, 4, 0, false, false)
		if err != nil {
			return err
		}

		var file *xlsx.File
		var cell *xlsx.Cell

		excelFileName := "UsersReg.xlsx"

		file, err = xlsx.OpenFile(excelFileName)
		if err != nil {
			return err
		}
		leader := me.Firstname + " " + me.Lastname + " " + me.Middlename
		for _, sheet := range file.Sheets {
			cell = AddCell(sheet, 0, 0)
			cell.Value = "ФИО"
			cell = AddCell(sheet, 0, 1)
			cell.Value = "Логин"
			cell = AddCell(sheet, 0, 2)
			cell.Value = "Пароль"
			cell = AddCell(sheet, 0, 3)
			cell.Value = "Компетенция"
			cell = AddCell(sheet, rc, 0)
			cell.Value = leader
			cell = AddCell(sheet, rc, 1)
			cell.Value = login
			cell = AddCell(sheet, rc, 2)
			cell.Value = password
			cell = AddCell(sheet, rc, 3)
			cell.Value = me.Competence
		}
		file.Save("UsersReg.xlsx")

		encpassword, err := encryptString(password)
		if err != nil {
			return err
		}

		rows, err := r.store.db.Query("SELECT MAX(id) FROM vkr.authorization")
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&aut_id)
			}
		}
		aut_id++

		if _, err := r.store.db.Exec(
			"INSERT INTO vkr.authorization (id, login, password) VALUES (?,?,?)",
			aut_id,
			login,
			encpassword,
		); err != nil {
			return err
		}

		rows, err = r.store.db.Query("SELECT MAX(id) FROM vkr.users")
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&me.ID)
			}
		}
		me.ID++

		rows, err = r.store.db.Query("SELECT id FROM vkr.role WHERE name = ? ",
			me.Role)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&role_id)
			}
		}

		rows, err = r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
			me.Competence)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&com_id)
			}
		}

		if _, err := r.store.db.Exec(
			"INSERT INTO vkr.users (id, firstname, lastname, middlename, year, role_id, authorization_id, competence_id) VALUES (?,?,?,?,?,?,?,?)",
			me.ID,
			me.Firstname,
			me.Lastname,
			me.Middlename,
			me.Year,
			role_id,
			aut_id,
			com_id,
		); err != nil {
			return err
		}
	}
	return err
}

func encryptString(s string) (string, error) {

	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", nil
	}

	return string(b), nil
}

func AddCell(sheet *xlsx.Sheet, row, col int) *xlsx.Cell {
	for row >= len(sheet.Rows) {
		sheet.AddRow()
	}
	for col >= len(sheet.Rows[row].Cells) {
		sheet.Rows[row].AddCell()
	}
	return sheet.Cell(row, col)
}

func (r *UserRepository) CreateTasks(ts *model.Tasks) error {
	var (
		com_id int
		id     int
	)
	rows, err := r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ?",
		ts.Competence)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}

	rows, err = r.store.db.Query("SELECT id FROM vkr.tasks WHERE jobnumber = ? and  year = ? and competence_id =?",
		ts.Jobnumber, ts.Year, com_id)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&id)
		}
	}

	if id != 0 {
		if _, err := r.store.db.Exec(
			"UPDATE vkr.tasks SET jobnumber = ?, aspect = ?, maxpoint = ?, year = ?, module = ?, competence_id = ? WHERE id = ?",
			ts.Jobnumber,
			ts.Aspect,
			ts.Maxpoint,
			ts.Year,
			ts.Module,
			com_id,
			id,
		); err != nil {
			return err
		}
	} else if id == 0 {
		rows, err = r.store.db.Query("SELECT MAX(id) FROM vkr.tasks")
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&ts.ID)
			}
		}
		ts.ID++
		rows, err = r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
			ts.Competence)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&com_id)
			}
		}

		if _, err := r.store.db.Exec(
			"INSERT INTO vkr.tasks (id,jobnumber, aspect, maxpoint, year, module, competence_id) VALUES (?,?,?,?,?,?,?)",
			ts.ID,
			ts.Jobnumber,
			ts.Aspect,
			ts.Maxpoint,
			ts.Year,
			ts.Module,
			com_id,
		); err != nil {
			return err
		}
	}
	return err
}

func (r *UserRepository) CreatePlayers(pl *model.Players) error {
	var (
		id     int
		com_id int
	)
	fmt.Print(pl)
	rows, err := r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
		pl.Competence)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}
	rows, err = r.store.db.Query("SELECT id FROM vkr.player WHERE  firstname = ? and  lastname = ? and middlename = ? and year = ? and competence_id = ?",
		pl.Firstname, pl.Lastname, pl.Middlename, pl.Year, com_id)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&id)
		}
	}
	if id != 0 {
		if _, err := r.store.db.Exec(
			"UPDATE vkr.player SET firstname = ?,  lastname= ?,  middlename= ?,  gender= ?,  dob = ?,  inclusion= ? ,leader = ?, year =?, competence_id = ?, workplace = ?  WHERE id = ?",
			pl.Firstname,
			pl.Lastname,
			pl.Middlename,
			pl.Gender,
			pl.Dob,
			pl.Inclusion,
			pl.Leader,
			pl.Year,
			com_id,
			pl.Workplace,
			id,
		); err != nil {
			return err
		}
	} else {
		rows, err := r.store.db.Query("SELECT MAX(id) FROM vkr.player")
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&pl.ID)
			}
		}
		pl.ID++

		var (
			com_id int
		)

		rows, err = r.store.db.Query("SELECT id FROM vkr.competence WHERE name = ? ",
			pl.Competence)
		if err != nil {
			return err
		} else {
			for rows.Next() {
				rows.Scan(&com_id)
			}
		}

		if _, err := r.store.db.Exec(
			"INSERT INTO vkr.player (id,firstname, lastname, middlename, gender, dob, inclusion, leader, year, competence_id, workplace) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
			pl.ID,
			pl.Firstname,
			pl.Lastname,
			pl.Middlename,
			pl.Gender,
			pl.Dob,
			pl.Inclusion,
			pl.Leader,
			pl.Year,
			com_id,
			pl.Workplace,
		); err != nil {
			return err
		}
	}
	return err
}

func (r *UserRepository) SelectTasksM(tsM model.TasksModel) ([]model.TasksModel, error) {
	var (
		com_id int
		u_id   int
	)
	arrTask := []model.TasksModel{}

	rows, err := r.store.db.Query("SELECT competence_id, id  FROM vkr.users  WHERE firstname = ? and lastname = ? and middlename = ?",
		tsM.Firstname, tsM.Lastname, tsM.Middlename)
	if err != nil {
		return arrTask, err
	} else {
		for rows.Next() {
			rows.Scan(&com_id, &u_id)
		}
	}
	defer rows.Close()
	// fmt.Println(com_id)
	rows, err = r.store.db.Query("Select jobnumber, aspect, maxpoint FROM vkr.tasks WHERE module = ? and competence_id = ? ",
		tsM.Module, com_id)
	if err != nil {
		return arrTask, err
	} else {
		for rows.Next() {
			err := rows.Scan(&tsM.Jobnumber, &tsM.Aspect, &tsM.Maxpoint)
			if err != nil {
				continue
			}
			arrTask = append(arrTask, tsM)
		}
	}

	return arrTask, err
}

func (r *UserRepository) SelectPlayers(tsM model.PlayersModel) ([]model.PlayersModel, error) {
	var (
		com_id int
	)
	arrPlayers := []model.PlayersModel{}

	rows, err := r.store.db.Query("SELECT competence_id FROM vkr.users  WHERE firstname = ? and lastname = ? and middlename = ?",
		tsM.Firstname, tsM.Lastname, tsM.Middlename)
	if err != nil {
		return arrPlayers, err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}
	defer rows.Close()

	//fmt.Println(com_id)
	leader := tsM.Firstname + " " + tsM.Lastname + " " + tsM.Middlename

	//fmt.Println(leader)

	rows, err = r.store.db.Query("Select firstname, lastname, middlename FROM vkr.player WHERE competence_id = ? and leader != ?",
		com_id, leader)
	if err != nil {
		return arrPlayers, err
	} else {
		for rows.Next() {
			err := rows.Scan(&tsM.FirstNameP, &tsM.LastnameP, &tsM.MiddlenameP)
			if err != nil {
				continue
			}
			arrPlayers = append(arrPlayers, tsM)
		}
	}
	return arrPlayers, err
}

func (r *UserRepository) SelectWork(pl *model.Players) error {

	rows, err := r.store.db.Query("Select workplace FROM vkr.player WHERE  firstname = ? and  lastname = ? and middlename = ?",
		pl.Firstname, pl.Lastname, pl.Middlename)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			err := rows.Scan(&pl.Workplace)
			if err != nil {
				continue
			}
		}
	}
	return err
}
func (r *UserRepository) CreateScoreSheet(ss model.Scoresheet) error {
	var (
		t_id   int
		u_id   int
		pl_id  int
		com_id int
		id     int
		idpr   int
	)

	rows, err := r.store.db.Query("Select MAX(id) FROM vkr.scoresheet")
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&id)
		}
	}
	id++
	rows, err = r.store.db.Query("Select id FROM vkr.player WHERE firstname = ? and lastname = ? and middlename = ?",
		ss.FirstnameP, ss.LastnameP, ss.MiddlenameP)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&pl_id)
		}
	}
	fmt.Println(pl_id)
	rows, err = r.store.db.Query("Select id, competence_id FROM vkr.users WHERE firstname = ? and lastname = ? and middlename = ?",
		ss.FirstName, ss.Lastname, ss.Middlename)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&u_id, &com_id)
		}
	}
	fmt.Println(u_id, com_id)
	rows, err = r.store.db.Query("Select id FROM vkr.tasks WHERE jobnumber = ? and  competence_id = ?",
		ss.Task, com_id)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&t_id)
		}
	}
	fmt.Println(t_id)
	rows, err = r.store.db.Query("SELECT id FROM vkr.scoresheet WHERE   pk = ? and   users_id = ? and tasks_id = ? and  player_id = ?",
		ss.Workplace, u_id, t_id, pl_id)
	if err != nil {
		return err
	} else {
		for rows.Next() {
			rows.Scan(&idpr)
		}
	}
	if idpr != 0 {
		if _, err := r.store.db.Exec(
			"UPDATE vkr.scoresheet SET  pk = ?,  setpoint = ?,  player_id = ?, tasks_id  = ?, users_id = ?  WHERE id = ?",
			ss.Workplace,
			ss.Setpoint,
			pl_id,
			t_id,
			u_id,
			idpr,
		); err != nil {
			return err
		}
	} else {
		if _, err := r.store.db.Exec("INSERT INTO vkr.scoresheet (id, pk ,setpoint ,player_id ,tasks_id ,users_id  ) VALUES (?,?,?,?,?,?)",
			id,
			ss.Workplace,
			ss.Setpoint,
			pl_id,
			t_id,
			u_id,
		); err != nil {
			return err
		}
	}
	return err
}

func (r *UserRepository) ListWinners(me *model.Aeg) ([]model.ListWinners, error) {
	var (
		total_score float64
		min_id      int
		max_idp     int
		max_idt     int
		jobnumber   string
		n           int
		setpoint    float64
		maxpoint    float64
		km          int
		kn          int
		i           int
		j           int
		// min         float64
		// max         float64
		count  float64
		temp   float64
		com_id int
	)
	list := []model.ListWinners{}
	listwin := model.ListWinners{}

	rows, err := r.store.db.Query("Select competence_id FROM vkr.users WHERE firstname = ? and lastname = ? and middlename = ?",
		me.Firstname, me.Lastname, me.Middlename)
	if err != nil {
		return list, err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}
	rows, err = r.store.db.Query("Select MIN(id), MAX(id) FROM vkr.player WHERE competence_id = ?", com_id)
	if err != nil {
		return list, err
	} else {
		for rows.Next() {
			rows.Scan(&min_id, &max_idp)
		}
	}
	for i = min_id; i <= max_idp; i++ {
		rows, err := r.store.db.Query("Select MIN(id), MAX(id) FROM vkr.tasks")
		if err != nil {
			return list, err
		} else {
			for rows.Next() {
				rows.Scan(&min_id, &max_idt)
			}
		}
		fmt.Println(i)
		for j = min_id; j <= max_idt; j++ {
			rows, err := r.store.db.Query("Select jobnumber FROM vkr.tasks WHERE id = ?", j)
			if err != nil {
				return list, err
			} else {
				for rows.Next() {
					rows.Scan(&jobnumber)
				}
			}
			fmt.Println(jobnumber)
			if strings.Contains(jobnumber, "о") == true {
				rows, err = r.store.db.Query("Select maxpoint FROM vkr.tasks WHERE id = ?", j)
				if err != nil {
					return list, err
				} else {
					for rows.Next() {
						rows.Scan(&maxpoint)
					}
				}
				// fmt.Println(maxpoint)
				arrtotalscore := []float64{}
				rows, err = r.store.db.Query("Select setpoint FROM vkr.scoresheet WHERE player_id = ? and tasks_id = ?", i, j)
				if err != nil {
					return list, err
				} else {
					for rows.Next() {
						rows.Scan(&setpoint)
						arrtotalscore = append(arrtotalscore, setpoint)
					}
				}
				fmt.Println(arrtotalscore)
				for y := 0; y < len(arrtotalscore); y++ {
					if arrtotalscore[y] == maxpoint {
						km++
					} else if arrtotalscore[y] == 0 {
						kn++
					}
				}
				// fmt.Println(km, kn)
				if km > kn {
					total_score += maxpoint
				} else if km < kn {
					total_score += 0
				}
				km = 0
				kn = 0
				fmt.Println(total_score)
				//fmt.Println(total_score)
			} else if strings.Contains(jobnumber, "с") == true {
				//fmt.Println(i)
				rows, err := r.store.db.Query("Select COUNT(id) FROM vkr.scoresheet WHERE player_id = ? and tasks_id = ?", i, j)
				if err != nil {
					return list, err
				} else {
					for rows.Next() {
						rows.Scan(&n)
					}
				}
				// fmt.Println(n)
				arrtotalscore := []float64{}
				rows, err = r.store.db.Query("Select setpoint FROM vkr.scoresheet WHERE player_id = ? and tasks_id = ?", i, j)
				if err != nil {
					return list, err
				} else {
					for rows.Next() {
						rows.Scan(&setpoint)
						arrtotalscore = append(arrtotalscore, setpoint)
					}
				}
				if len(arrtotalscore) == 0 {
					total_score += 0
				} else {
					sort.Float64s(arrtotalscore)

					n -= 1
					// fmt.Println(n)
					arrsrez := arrtotalscore[1:n]
					count = 0
					for _, value := range arrsrez {
						temp += value
						count++
					}
					fmt.Println(arrsrez)
					fmt.Println(temp)
					fmt.Println(count)
					total_score += temp / count
					total_score = (math.Floor(total_score*100) / 100)
					temp = 0
					fmt.Println(total_score)
				}
			}
		}
		rows, err = r.store.db.Query("Select firstname, lastname, middlename FROM vkr.player WHERE  id = ?", i)
		if err != nil {
			return list, err
		} else {
			for rows.Next() {
				rows.Scan(&listwin.FirstName, &listwin.Lastname, &listwin.Middlename)
			}
		}

		total_score = (math.Floor(total_score*100) / 100)
		listwin.Totalscore += total_score
		total_score = 0
		list = append(list, listwin)
		listwin.Totalscore = 0
	}

	return list, err
}

//FindByLogin
func (r *UserRepository) FindByLogin(login string) (*model.User, error) {
	u := &model.User{}
	var(com_id int)
	if err := r.store.db.QueryRow(
		"SELECT id, login, Password FROM vkr.authorization  WHERE login = ?",
		login,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	if err := r.store.db.QueryRow(
		"SELECT FirstName, LastName,Middlename, role_id FROM vkr.users WHERE authorization_id = ?",
		&u.ID,
	).Scan(
		&u.FirstName,
		&u.Lastname,
		&u.Middlename,
		&u.Role,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	rows, err := r.store.db.Query(
	"SELECT competence_id FROM vkr.users WHERE Firstname = ? and  Lastname = ? and  Middlename = ?",
	&u.FirstName,&u.Lastname,&u.Middlename,
	)
	if err != nil{
		return nil, err
	} else {
		for rows.Next() {
			rows.Scan(&com_id)
		}
	}

	rows, err = r.store.db.Query("SELECT MAX(module) FROM vkr.tasks WHERE competence_id = ?", 
	&com_id,)
	if err != nil{
		return nil, err
	} else{
		for rows.Next(){
			rows.Scan(&u.Module)
		}
	}
	

	if err := r.store.db.QueryRow(
		"SELECT name FROM vkr.role WHERE id = ?",
		&u.Role,
	).Scan(
		&u.Role_Name,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

//Find
func (r *UserRepository) Find(id int) (*model.User, error) {
	u := &model.User{}

	if err := r.store.db.QueryRow(
		"SELECT id, login, Password FROM vkr.authorization WHERE id = ?",
		id,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Password,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	if err := r.store.db.QueryRow(
		"SELECT role_id FROM vkr.users WHERE authorization_id = ?",
		&u.ID,
	).Scan(
		&u.Role,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	if err := r.store.db.QueryRow(
		"SELECT name FROM vkr.role WHERE id = ?",
		&u.Role,
	).Scan(
		&u.Role_Name,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
