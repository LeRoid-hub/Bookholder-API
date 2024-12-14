package database

type DB struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

func SetEnv(env map[string]string) DB {
	var db DB

	db.Host = env["DB_HOST"]
	db.User = env["DB_USER"]
	db.Password = env["DB_PASSWORD"]
	db.Name = env["DB_NAME"]
	db.Port = env["DB_PORT"]

	return db
}

func connect() {

}

func createTables() {

}

func updateTables() {

}

func NewAccount() {

}

func UpdateAccount() {

}

func DeleteAccount() {

}

func GetAccount() {

}

func NewTransaction() {

}

func UpdateTransaction() {

}

func DeleteTransaction() {

}

func GetTransactions() {

}

func NewUser() {

}

func UpdateUser() {

}

func DeleteUser() {

}

func GetUser() {

}
