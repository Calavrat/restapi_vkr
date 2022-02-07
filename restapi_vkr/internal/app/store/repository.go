package store

import "github.com/Calavrat/http-rest-api/internal/app/model"

//UserRepository
type UserRepository interface {
	Create(*model.User) error
	CreateUsers(*model.Aeg, int) error
	CreateUsersG(*model.Aeg, int) error
	CreatePlayers(*model.Players) error
	CreateTasks(*model.Tasks) error
	CreateScoreSheet(model.Scoresheet) error
	Find(int) (*model.User, error)
	FindByLogin(string) (*model.User, error)
	SelectComT(string) ([]model.Competence, error)
	SelectTasksM(model.TasksModel) ([]model.TasksModel, error)
	SelectPlayers(model.PlayersModel) ([]model.PlayersModel, error)
	ListWinners(me *model.Aeg) ([]model.ListWinners, error)
	SelectWork(pl *model.Players) error
}
