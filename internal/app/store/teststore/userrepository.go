package teststore

import (
	"github.com/Calavrat/http-rest-api/internal/app/model"
	"github.com/Calavrat/http-rest-api/internal/app/store"
)

type UserRepository struct {
	store      *Store
	users      map[int]*model.User
	competence map[int]*model.Competence
	mainexpert map[int]*model.Aeg
	players    map[int]*model.Players
	tasks      map[int]*model.Tasks
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	if err := u.BeforeCreate(); err != nil {
		return err
	}
	u.ID = len(r.users) + 1
	r.users[u.ID] = u
	return nil
}

func (r *UserRepository) CreateUsers(me *model.Aeg, rc int) error {
	me.ID = len(r.mainexpert) + 1
	r.mainexpert[me.ID] = me
	return nil
}

func (r *UserRepository) CreateUsersG(me *model.Aeg, rc int) error {
	me.ID = len(r.mainexpert) + 1
	r.mainexpert[me.ID] = me
	return nil
}

func (r *UserRepository) CreatePlayers(pl *model.Players) error {
	pl.ID = len(r.players) + 1
	r.players[pl.ID] = pl
	return nil
}

func (r *UserRepository) CreateTasks(ts *model.Tasks) error {
	ts.ID = len(r.tasks) + 1
	r.tasks[ts.ID] = ts
	return nil
}

func (r *UserRepository) CreateScoreSheet(model.Scoresheet) error {
	return store.ErrRecordNotFound
}

func (r *UserRepository) FindByLogin(login string) (*model.User, error) {
	for _, u := range r.users {
		if u.Login == login {
			return u, nil
		}
	}
	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	u, ok := r.users[id]

	if !ok {
		return nil, store.ErrRecordNotFound
	}
	return u, nil
}

func (r *UserRepository) SelectComT(Type string) ([]model.Competence, error) {

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) SelectTasksM(tsM model.TasksModel) ([]model.TasksModel, error) {

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) SelectPlayers(tsM model.PlayersModel) ([]model.PlayersModel, error) {

	return nil, store.ErrRecordNotFound
}
func (r *UserRepository) ListWinners(me *model.Aeg) ([]model.ListWinners, error) {

	return nil, store.ErrRecordNotFound
}
func (r *UserRepository) SelectWork(pl *model.Players) error {

	return store.ErrRecordNotFound
}
