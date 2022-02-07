package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	//"io/ioutil"
	"github.com/Calavrat/http-rest-api/internal/app/model"
	"github.com/Calavrat/http-rest-api/internal/app/store"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	sessionName        = "Calavrat"
	ctxKeyUser  ctxKey = iota
	ctxKeyRequestID
)

var (
	errIncorrectLoginOrPassword = errors.New("Incorrect login or password") // ошибка если пользователь ввел не верный пароль или логин
	errNotAuthenticated         = errors.New("Not authenticated")           //оштбка для неавторизованного пользователя
	errNotDown                  = errors.New("Error retrieving file from  form-data")
	users                       = "Эксперты импортированы"
	players                     = "Участники импортированы"
	task                        = "Задания импортированы"
)

type ctxKey int8

type server struct { // структура с зависимостями
	router       *mux.Router
	logger       *logrus.Logger
	store        store.Store
	sessionStore sessions.Store
}

func newServer(store store.Store, sessionStore sessions.Store) *server { //функция конструктор на выходе получаем сервер
	s := &server{ // инициализируем перменные
		router:       mux.NewRouter(),
		logger:       logrus.New(),
		store:        store,
		sessionStore: sessionStore,
	}
	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r) //делегируем функцию ServeHTTP роутеру
}

func (s *server) configureRouter() {
	//функция создает маршруты
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.HandleFunc("/users", s.handleUsersCreate()).Methods("POST")
	s.router.HandleFunc("/sessions", s.handleSessionsCreate()).Methods("POST")
	s.router.HandleFunc("/upload", s.handleDownFile()).Methods("POST")
	s.router.HandleFunc("/selectcomt", s.handleSelectComT()).Methods("POST")
	s.router.HandleFunc("/mainexpert", s.handleExpertCreate()).Methods("POST")
	s.router.HandleFunc("/selecttasksm", s.handleSelectTasks()).Methods("POST")
	s.router.HandleFunc("/scoresheet", s.handleScoresheet()).Methods("POST")
	s.router.HandleFunc("/listwinners", s.handleListWinners()).Methods("POST")
	s.router.HandleFunc("/selectwork", s.handleSelectWork()).Methods("POST")

	private := s.router.PathPrefix("/private").Subrouter()
	private.Use(s.authenticateUser)
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-RequestID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})

		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Infof(
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start))

	})
}

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
}

//handleUsersCreate
func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		//описываем те поля оторые ожидаем от клиента
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Login:    req.Login,
			Password: req.Password,
		}

		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		defer r.Body.Close()

		//u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

//handleSessionCreate
func (s *server) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByLogin(req.Login)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		session.Values["user_id"] = u.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		defer r.Body.Close()
		u.Sanitize()
		s.respond(w, r, http.StatusOK, u)
	}
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		u, err := s.store.User().Find(id.(int))
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		u.Sanitize()
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})

}

func (s *server) handleSelectComT() http.HandlerFunc {
	type request struct {
		Type string `json:"type"`
	}

	type mapreq struct {
		MapToEncode []model.Competence `json:"map"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		mr := &mapreq{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		Type := req.Type
		defer r.Body.Close()
		comT, err := s.store.User().SelectComT(Type)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
		}

		for _, c := range comT {
			c.SanitizeC()
			mr.MapToEncode = append(mr.MapToEncode, c)
		}
		defer r.Body.Close()
		s.respond(w, r, http.StatusCreated, mr)
	}
}

func (s *server) handleExpertCreate() http.HandlerFunc {
	type request struct {
		Firstname  string `json:"firstname"`
		Lastname   string `json:"lastname"`
		Middlename string `json:"middlename"`
		Role       string `json:"role"`
		Competence string `json:"competence"`
		Year       string `json:"year"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		me := &model.Aeg{
			Firstname:  req.Firstname,
			Lastname:   req.Lastname,
			Middlename: req.Middlename,
			Role:       req.Role,
			Competence: req.Competence,
			Year:       req.Year,
		}
		me.Year += ".10.01"

		excelFileName := "/root/Projects/restapi_vkr/cmd/apiserver/MainUsersReg.xlsx"
		xlFile, err := xlsx.OpenFile(excelFileName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		idx := 0
		for _, sheet := range xlFile.Sheets {
			for idx, _ = range sheet.Rows {
				idx++
			}
		}
		
		if err := s.store.User().CreateUsersG(me, idx); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		defer r.Body.Close()

		s.respond(w, r, http.StatusCreated, me)
	}
}

func (s *server) handleDownFile() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			status int
			err    error
		)
		defer func() {
			if nil != err {
				http.Error(w, err.Error(), status)
			}
		}()
		// parse request
		// const _24K = (1 << 20) * 24
		if err = r.ParseMultipartForm(32 << 20); nil != err {
			status = http.StatusInternalServerError
			return
		}
		fmt.Println("No memory problem")
		for _, fheaders := range r.MultipartForm.File {
			for _, hdr := range fheaders {
				// open uploaded
				var infile multipart.File
				if infile, err = hdr.Open(); nil != err {
					status = http.StatusInternalServerError
					return
				}
				// open destination
				var outfile *os.File
				if outfile, err = os.Create("/root/Projects/restapi_vkr/cmd/apiserver/" + hdr.Filename); nil != err {
					status = http.StatusInternalServerError
					return
				}
				// 32K buffer copy
				var written int64
				if written, err = io.Copy(outfile, infile); nil != err {
					status = http.StatusInternalServerError
					return
				}
				win := []byte("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written)))
				logrus.Printf("%q\n", win)
				defer r.Body.Close()
				if hdr.Filename == "users.xlsx" {
					var (
						dd string
						mm string
						gg string
					)
					excelFileName := "/root/Projects/restapi_vkr/cmd/apiserver/users.xlsx"
					xlFile, err := xlsx.OpenFile(excelFileName)
					if err != nil {
						s.error(w, r, http.StatusInternalServerError, err)
						return
					}
					me := &model.Aeg{}
						

					for _, sheet := range xlFile.Sheets {
						for idx, row := range sheet.Rows {
							if idx == 0 {
								for id, cell := range row.Cells {
									text := cell.String()
									if id == 3 {
										me.Year = text
									}
								}
							} else if idx == 1 {
								for id, cell := range row.Cells {
									text := cell.String()
									if id == 3 {
										me.Competence = text
									}
								}
							} else {
								continue
							}
						}
					}
					arr := make([]string, 3)
					arr = strings.Split(me.Year, " ")
					dd = arr[0]
					mm = arr[1]
					gg = arr[2]

					switch mm {
					case "января":
						mm = "1"
					case "февраля":
						mm = "2"
					case "марта":
						mm = "3"
					case "апреля":
						mm = "4"
					case "мая":
						mm = "5"
					case "июня":
						mm = "6"
					case "июля":
						mm = "7"
					case "августа":
						mm = "8"
					case "сентября":
						mm = "9"
					case "октября":
						mm = "10"
					case "ноября":
						mm = "11"
					case "декабря":
						mm = "12"
					}

					me.Year = gg + "." + mm + "." + dd
					
					excelFileName = "/root/Projects/restapi_vkr/cmd/apiserver/UsersReg.xlsx"
					xlFile, err = xlsx.OpenFile(excelFileName)
					if err != nil {
						s.error(w, r, http.StatusInternalServerError, err)
						return
					}
					idxp := 0
					for _, sheet := range xlFile.Sheets {
						for idxp, _ = range sheet.Rows {
							idxp++
						}
					}
					logrus.Println(idxp)

					excelFileName = "/root/Projects/restapi_vkr/cmd/apiserver/users.xlsx"
					xlFile, err = xlsx.OpenFile(excelFileName)
					if err != nil {
						s.error(w, r, http.StatusInternalServerError, err)
						return
					}
					for _, sheet := range xlFile.Sheets {
						for idx, row := range sheet.Rows {
							if idx < 3 {
								continue
							}
							for id, cell := range row.Cells {
								text := cell.String()
								re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
								re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
								final := re_leadclose_whtsp.ReplaceAllString(text, "")
								final = re_inside_whtsp.ReplaceAllString(final, " ")
								switch id {
								case 0:
									me.Firstname = final

								case 1:
									me.Lastname = final
									
								case 2:
									me.Middlename = final
									
								case 3:
									me.Role = final
									
								}
							}
							
							if err := s.store.User().CreateUsers(me, idxp); err != nil {
								s.error(w, r, http.StatusUnprocessableEntity, err)
								return
							}
							idxp++
						}
					}
					s.respond(w, r, http.StatusCreated, users)
				} else if hdr.Filename == "tasks.xlsx" {
					var (
						dd string
						mm string
						gg string
					)
					exselFileName := "/root/Projects/restapi_vkr/cmd/apiserver/tasks.xlsx"
					xlFile, err := xlsx.OpenFile(exselFileName)
					if err != nil {
						s.error(w, r, http.StatusInternalServerError, err)
						return
					}
					ts := &model.Tasks{}
					for _, sheet := range xlFile.Sheets {
						for idx, row := range sheet.Rows {
							if idx == 0 {
								for id, cell := range row.Cells {
									text := cell.String()
									if id == 2 {
										ts.Year = text
									}
								}
							} else if idx == 1 {
								for id, cell := range row.Cells {
									text := cell.String()
									if id == 2 {
										ts.Competence = text
									}
								}
							} else {
								continue
							}
						}
					}
					arr := make([]string, 3)
					arr = strings.Split(ts.Year, " ")
					dd = arr[0]
					mm = arr[1]
					gg = arr[2]
					switch mm {
					case "января":
						mm = "1"
					case "февраля":
						mm = "2"
					case "марта":
						mm = "3"
					case "апреля":
						mm = "4"
					case "мая":
						mm = "5"
					case "июня":
						mm = "6"
					case "июля":
						mm = "7"
					case "августа":
						mm = "8"
					case "сентября":
						mm = "9"
					case "октября":
						mm = "10"
					case "ноября":
						mm = "11"
					case "декабря":
						mm = "12"
					}

					ts.Year = gg + "." + mm + "." + dd
					for _, sheet := range xlFile.Sheets {
						for idx, row := range sheet.Rows {
							if idx < 3 {
								continue
							}
							for id, cell := range row.Cells {
								text := cell.String()
								re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
								re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
								final := re_leadclose_whtsp.ReplaceAllString(text, "")
								final = re_inside_whtsp.ReplaceAllString(final, " ")
								if id == 0 {
									switch {
									case strings.Contains(final, "A") == true:
										ts.Module = 1
									case strings.Contains(final, "B") == true:
										ts.Module = 2
									case strings.Contains(final, "C") == true:
										ts.Module = 3
									case strings.Contains(final, "D") == true:
										ts.Module = 4
									case strings.Contains(final, "E") == true:
										ts.Module = 5
									case strings.Contains(final, "F") == true:
										ts.Module = 6
									case strings.Contains(final, "G") == true:
										ts.Module = 7
									}
								}
								switch id {
								case 0:
									ts.Jobnumber = final
								case 1:
									ts.Aspect = final
								case 2:
									i, err := strconv.Atoi(final)
									if err != nil {
										s.error(w, r, http.StatusUnprocessableEntity, err)
									}
									ts.Maxpoint = i
								}
							}
							if err := s.store.User().CreateTasks(ts); err != nil {
								s.error(w, r, http.StatusUnprocessableEntity, err)
								return
							}
							ts.SanitizeTasks()

						}
					}
					s.respond(w, r, http.StatusCreated, task)
				} else if hdr.Filename == "players.xlsx" {
					var (
						dd string
						mm string
						gg string
					)

					exselFileName := "/root/Projects/restapi_vkr/cmd/apiserver/players.xlsx"
					xlFile, err := xlsx.OpenFile(exselFileName)
					if err != nil {
						s.error(w, r, http.StatusInternalServerError, err)
						return
					}

					pl := &model.Players{}

					for _, sheet := range xlFile.Sheets {
						for idx, row := range sheet.Rows {
							if idx == 0 {
								for id, cell := range row.Cells {
									text := cell.String()
									if id == 7 {
										pl.Year = text
									}
								}
							} else if idx == 1 {
								for id, cell := range row.Cells {
									text := cell.String()
									if id == 7 {
										pl.Competence = text
									}
								}
							} else {
								continue
							}
						}
					}
					arr := make([]string, 3)
					arr = strings.Split(pl.Year, " ")
					dd = arr[0]
					mm = arr[1]
					gg = arr[2]

					switch mm {
					case "января":
						mm = "1"
					case "февраля":
						mm = "2"
					case "марта":
						mm = "3"
					case "апреля":
						mm = "4"
					case "мая":
						mm = "5"
					case "июня":
						mm = "6"
					case "июля":
						mm = "7"
					case "августа":
						mm = "8"
					case "сентября":
						mm = "9"
					case "октября":
						mm = "10"
					case "ноября":
						mm = "11"
					case "декабря":
						mm = "12"
					}
					pl.Year = gg + "." + mm + "." + dd

					for _, sheet := range xlFile.Sheets {
						for idx, row := range sheet.Rows {
							if idx < 3 {
								continue
							}
							for id, cell := range row.Cells {
								text := cell.String()
								switch id {
								case 0:
									pl.Firstname = text
								case 1:
									pl.Lastname = text
								case 2:
									pl.Middlename = text
								case 3:
									pl.Gender = text
								case 4:
									pl.Dob = text
								case 5:
									pl.Inclusion = text
								case 6:
									pl.Leader = text
								case 7:
									pl.Workplace = text
								}
							}
							arr = strings.Split(pl.Dob, "-")
							dd = arr[0]
							mm = arr[1]
							gg = arr[2]
							fmt.Println(dd, mm, "20"+gg)
							pl.Dob = "20" + gg + "." + mm + "." + dd
							fmt.Println(pl.Dob)
							if err := s.store.User().CreatePlayers(pl); err != nil {
								s.error(w, r, http.StatusUnprocessableEntity, err)
								return
							}

						}
					}
					s.respond(w, r, http.StatusCreated, players)
				}
				// w.Write([]byte("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written))))
			}
		}
	}
}

func (s *server) handleSelectTasks() http.HandlerFunc {
	type request struct {
		Firstname  string `json:"firstname"`
		Lastname   string `json:"lastname"`
		Middlename string `json:"middlename"`
		Module     int    `json:"module"`
	}
	type mapreq struct {
		MapToEncode  []model.TasksModel   `json:"task"`
		MapToPlayers []model.PlayersModel `json:"players"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		mreq := &mapreq{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		tsM := model.TasksModel{
			Firstname:  req.Firstname,
			Lastname:   req.Lastname,
			Middlename: req.Middlename,
			Module:     req.Module,
		}

		pl := model.PlayersModel{
			Firstname:  req.Firstname,
			Lastname:   req.Lastname,
			Middlename: req.Middlename,
			Module:     req.Module,
		}
		// fmt.Println(tsM)
		// fmt.Println(pl)
		arrTask, err := s.store.User().SelectTasksM(tsM)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
		}

		for _, t := range arrTask {
			t.SanitizeTasksM()
			mreq.MapToEncode = append(mreq.MapToEncode, t)
		}

		arrPlayers, err := s.store.User().SelectPlayers(pl)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
		}

		for _, p := range arrPlayers {
			p.SanitizePlayers()
			mreq.MapToPlayers = append(mreq.MapToPlayers, p)
		}
		defer r.Body.Close()
		s.respond(w, r, http.StatusOK, mreq)
	}

}
func (s *server) handleSelectWork() http.HandlerFunc {
	type request struct {
		Firstname  string `json:"firstname"`
		Lastname   string `json:"lastname"`
		Middlename string `json:"middlename"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		pl := &model.Players{
			Firstname:  req.Firstname,
			Lastname:   req.Lastname,
			Middlename: req.Middlename,
		}

		err := s.store.User().SelectWork(pl)
		if err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
		}
		defer r.Body.Close()
		pl.SanitizePlayersW()
		s.respond(w, r, http.StatusOK, pl)
	}
}
func (s *server) handleScoresheet() http.HandlerFunc {
	type Request struct {
		FirstnameP  string `json:"firstnamep"`
		LastnameP   string `json:"lastnamep"`
		MiddlenameP string `json:"middlenamep"`
		FirstName   string `json:"firstname"`
		Lastname    string `json:"lastname"`
		Middlename  string `json:"middlename"`
		Task        string `json:"task"`
		Workplace   string `json:"workplace"`
		Setpoint    string `json:"setpoint"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &Request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
		}

		st, err := strconv.ParseFloat(req.Setpoint, 64)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
		}
		logrus.Println(req)
		ss := model.Scoresheet{
			FirstnameP:  req.FirstnameP,
			LastnameP:   req.LastnameP,
			MiddlenameP: req.MiddlenameP,
			FirstName:   req.FirstName,
			Lastname:    req.Lastname,
			Middlename:  req.Middlename,
			Task:        req.Task,
			Workplace:   req.Workplace,
			Setpoint:    st,
		}

		logrus.Println(ss)

		if err := s.store.User().CreateScoreSheet(ss); err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
		}
		s.respond(w, r, http.StatusOK, ss)

	}
}

func (s *server) handleListWinners() http.HandlerFunc {
	type mapreq struct {
		MapToEncode []model.ListWinners `json:"win"`
	}
	type reqme struct {
		Firstname  string `json:"firstname"`
		Lastname   string `json:"lastname"`
		Middlename string `json:"middlename"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &mapreq{}
		mereq := &reqme{}

		if err := json.NewDecoder(r.Body).Decode(mereq); err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
		}
		me := &model.Aeg{
			Firstname:  mereq.Firstname,
			Lastname:   mereq.Lastname,
			Middlename: mereq.Middlename,
		}
		fmt.Println(mereq)
		arrWin, err := s.store.User().ListWinners(me)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
		}
		//fmt.Println(arrWin)
		// arrWin[0].Totalscore += 1.35
		arrWin[0].Totalscore = (math.Floor(arrWin[0].Totalscore*100) / 100)
		// arrWin[4].Totalscore += 0.33
		// arrWin[3].Totalscore -= 0.33
		sort.SliceStable(arrWin, func(i, j int) bool {
			return arrWin[i].Totalscore > arrWin[j].Totalscore
		})

		arr := []model.ListWinners{}
		o := 0
		k := 0
		count := 0
		g := 0
		for i := 0; i < len(arrWin); i++ {
			//fmt.Println(count)
			if count > 0 {
				k = i
			} else if count == 0 {
				k = i
				k++
			}
			for j := i + 1; j < len(arrWin); j++ {
				if arrWin[j].Totalscore == arrWin[i].Totalscore {
					arrWin[j].Place = k
					arrWin[i].Place = k
					arr = append(arr, arrWin[i])
					fmt.Println(arrWin[i], arrWin[j])
					i++
					fmt.Println(arrWin[i], arrWin[j])
					o = j
					count++
					g = k
				} else {
					arrWin[i].Place = k
				}
			}
			arr = append(arr, arrWin[i])
		}
		fmt.Println(arrWin)
		fmt.Println(o, g)
		n := len(arrWin)
		arrsrez9 := arr[o+1 : n]
		arrsrez1 := arr[0 : o+1]
		for i := 0; i < len(arrsrez9); i++ {
			//		fmt.Println(g)
			g++
			arrsrez9[i].Place = g
			arrsrez1 = append(arrsrez1, arrsrez9[i])
		}

		// fmt.Println(arrsrez9)
		// fmt.Println(arrsrez1)
		for _, value := range arrsrez1 {
			req.MapToEncode = append(req.MapToEncode, value)
		}
		s.respond(w, r, http.StatusOK, req)

	}
}
func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) { //помощник чтобы выводить информацию об ошибках
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data) //записывем в w data а data записана информация об ошибках
	}
}
