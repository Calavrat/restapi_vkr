//Этот модуль нужен для управления зависимостями в GO
// Это когда ваша программа зависит именно от тех версий библиотек, от которых зависит.'
// И какой-нибудь другой разработчик, взяв ваш код и попытавшись его собрать, получит те же самые зависимости
module github.com/Calavrat/http-rest-api

go 1.13

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.1.0
	github.com/BurntSushi/toml v0.3.1
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/lib/pq v1.3.0
	github.com/sethvargo/go-password v0.1.3
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/tealeg/xlsx v1.0.5
	golang.org/x/crypto v0.0.0-20200109152110-61a87790db17
)
