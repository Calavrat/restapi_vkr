package teststore

import (
	"github.com/Calavrat/http-rest-api/internal/app/model"
	"github.com/Calavrat/http-rest-api/internal/app/store"
)

type Store struct {
	userRepository *UserRepository
}

func New() *Store {
	return &Store{}
}

//User
func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store:      s,
		users:      make(map[int]*model.User),
		competence: make(map[int]*model.Competence),
		mainexpert: make(map[int]*model.Aeg),
		players:    make(map[int]*model.Players),
		tasks:      make(map[int]*model.Tasks),
	}
	return s.userRepository
}
