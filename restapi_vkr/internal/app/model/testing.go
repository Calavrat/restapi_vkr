package model

import "testing"

func TestUser(t *testing.T) *User {
	return &User{
		Login:    "Admin!",
		Password: "Admin010101",
	}

}
func TestCompetence(t *testing.T) *Competence {
	return &Competence{
		Id:   1,
		Name: "Обработка текста",
		Type: "IT-ТЕХНОЛОГИИ",
	}
}
func TestMainExpert(t *testing.T) *Aeg {
	return &Aeg{
		Firstname:  "Петров",
		Lastname:   "Петр",
		Middlename: "Акылбекович",
		Role:       "Главный Эксперт",
		Competence: "Обработка текста",
	}
}

func TestPlayers(t *testing.T) *Players {
	return &Players{
		Firstname:  "Шляпов",
		Lastname:   "Андрей",
		Middlename: "Сергеевич",
		Gender:     "Мужской",
		Dob:        "1999.06.29",
		Year:       "2001.01.01",
		Inclusion:  "2 группа",
		Competence: "Обработка текста",
	}
}

func TestTasks(t *testing.T) *Tasks {
	return &Tasks{
		ID:         1,
		Jobnumber:  "A1О01",
		Aspect:     "Задание1",
		Maxpoint:   3,
		Module:     1,
		Competence: "Веб-Дизайн",
	}
}
