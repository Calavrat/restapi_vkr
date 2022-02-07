package teststore_test

import (
	"testing"

	"github.com/Calavrat/http-rest-api/internal/app/model"
	"github.com/Calavrat/http-rest-api/internal/app/store"
	"github.com/Calavrat/http-rest-api/internal/app/store/teststore"
	"github.com/stretchr/testify/assert"
)

func TestUserRepositore_Create(t *testing.T) {

	s := teststore.New()
	u := model.TestUser(t)

	assert.NoError(t, s.User().Create(u))
	assert.NotNil(t, u)
}

func TestUserRepositore_CreatePlayers(t *testing.T) {
	s := teststore.New()
	pl := model.TestPlayers(t)
	assert.NoError(t, s.User().CreatePlayers(pl))
	assert.NotNil(t, pl)
}
func TestUserRepositore_Create_MainExpert(t *testing.T) {
	s := teststore.New()
	me := model.TestMainExpert(t)
	rc := 1
	assert.NoError(t, s.User().CreateUsers(me, rc))
	assert.NotNil(t, me)
}
func TestUserRepositore_FindByLogin(t *testing.T) {

	s := teststore.New()
	login := "Expert"
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

	s := teststore.New()
	u1 := model.TestUser(t)
	s.User().Create(u1)

	u2, err := s.User().Find(u1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, u2)
}
