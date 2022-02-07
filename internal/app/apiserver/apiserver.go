package apiserver

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Calavrat/http-rest-api/internal/app/store/sqlstore"
	"github.com/gorilla/sessions"
)

func Start(config *Config) error { //c помощью этой функции запускаем HTTP сервер и подключаемся к БД
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	defer db.Close()

	store := sqlstore.New(db)
	sessionStore := sessions.NewCookieStore([]byte(config.SessionKey))
	srw := newServer(store, sessionStore)
	fmt.Println("Server is listening...")
	return http.ListenAndServe(config.BindAddr, srw)
}

func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}
