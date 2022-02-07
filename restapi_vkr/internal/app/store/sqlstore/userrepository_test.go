package sqlstore_test

import (
	"testing"

	"github.com/Calavrat/http-rest-api/internal/app/model"
	"github.com/Calavrat/http-rest-api/internal/app/store"
	"github.com/Calavrat/http-rest-api/internal/app/store/sqlstore"
	"github.com/stretchr/testify/assert"
)

func TestUserRepositore_Create(t *testing.T) {

	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("User")

	s := sqlstore.New(db)
	u := model.TestUser(t)

	assert.NoError(t, s.User().Create(u))
	assert.NotNil(t, u)
}
func TestUserRepositore_CreateUsers(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("MainExpert")

	s := sqlstore.New(db)
	me := model.TestMainExpert(t)
	id := 1
	assert.NoError(t, s.User().CreateUsers(me, id))
	assert.NotNil(t, me)
}
func TestUserRepositore_CreatePlayers(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("Players")

	s := sqlstore.New(db)
	pl := model.TestPlayers(t)

	assert.NoError(t, s.User().CreatePlayers(pl))
	assert.NotNil(t, pl)
}
func TestUserRepositore_CreateTasks(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("Tasks")

	s := sqlstore.New(db)
	ts := model.TestTasks(t)

	assert.NoError(t, s.User().CreateTasks(ts))
	assert.NotNil(t, ts)
}
func TestUserRepositore_FindByLogin(t *testing.T) {

	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("Users")

	s := sqlstore.New(db)
	login := "Admin!"
	_, err := s.User().FindByLogin(login)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	u := model.TestUser(t)
	u.Login = login
	s.User().Create(u)

	u, err = s.User().FindByLogin(login)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserRepositore_Find(t *testing.T) {

	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("Users")

	s := sqlstore.New(db)
	u1 := model.TestUser(t)
	s.User().Create(u1)

	u2, err := s.User().Find(u1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, u2)
}
