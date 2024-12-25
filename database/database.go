package database

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

var (
	database DB
)

const (
	SqlFile = "./database/bookholder.sql"
)

func SetEnv(env map[string]string) *DB {
	var db DB

	db.Host = env["DB_HOST"]
	db.User = env["DB_USER"]
	db.Password = env["DB_PASSWORD"]
	db.Name = env["DB_NAME"]
	db.Port = env["DB_PORT"]

	database = db

	checkDatabase()

	return &db
}

func New() (*sql.DB, error) {
	conn, err := sql.Open("pgx", connect())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func connect() string {
	return "postgres://" + database.User + ":" + database.Password + "@" + database.Host + ":" + database.Port + "/" + database.Name + "?sslmode=disable"
	//return "user=" + database.User + " password=" + database.Password + " host=" + database.Host + " dbname=" + database.Name + " sslmode=disable"
}

func checkDatabase() {
	conn, err := sql.Open("pgx", "postgres://"+database.User+":"+database.Password+"@"+database.Host+":"+database.Port+"/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = conn.QueryRow("SELECT datname FROM pg_database WHERE datname = $1", database.Name).Scan(&database.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			createDatabase()
		} else {
			panic(err)
		}
	}
}

func createDatabase() {
	conn, err := sql.Open("pgx", "postgres://"+database.User+":"+database.Password+"@"+database.Host+":"+database.Port+"/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Exec("CREATE DATABASE " + database.Name)
	if err != nil {
		panic(err)
	}

	createTables()
}

func checkTables() {

}

func createTables() {
	sqlFile, err := os.ReadFile(SqlFile)
	if err != nil {
		panic(err)
	}

	sqlStatements := strings.Split(string(sqlFile), ";")

	conn, err := sql.Open("pgx", connect())

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	for _, statement := range sqlStatements {
		_, err := conn.Exec(statement)
		if err != nil {
			panic(err)
		}
	}

}

func updateTables() {

}

func existAccount(database *sql.DB, id int) (bool, error) {
	err := database.QueryRow("SELECT id FROM accounts WHERE id = $1", id).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil

}

func NewAccount(database *sql.DB, account Account) error {
	exists, err := existAccount(database, int(account.ID))
	if err != nil {
		return err
	}
	if exists {
		return errors.New("account already exists")
	}

	_, err = database.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account.ID, account.Name, account.Kind)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAccount(database *sql.DB, account Account) error {
	_, err := database.Exec("UPDATE accounts SET name = $1, kind = $2 WHERE id = $3", account.Name, account.Kind, account.ID)
	if err != nil {
		return err
	}
	return nil

}

func DeleteAccount(database *sql.DB, id int) error {
	_, err := database.Exec("DELETE FROM accounts WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil

}

func GetAccount(database *sql.DB, id int) (Account, error) {
	var account Account
	err := database.QueryRow("SELECT * FROM accounts WHERE id = $1", id).Scan(&account.ID, &account.Name, &account.Kind)
	if err != nil {
		return account, err
	}
	return account, nil

}

func NewTransaction(database *sql.DB, transaction Transaction) error {
	_, err := database.Exec("INSERT INTO transactions (amount, debit, offset_account, account, time, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction.Amount, transaction.Debit, transaction.OffsetAccount, transaction.Account, transaction.Date, transaction.Description)
	if err != nil {
		return err
	}
	return nil

}

func UpdateTransaction(database *sql.DB, transaction Transaction) error {
	_, err := database.Exec("UPDATE transactions SET amount = $1, debit = $2, offset_account = $3, account = $4, time = $5, description = $6 WHERE id = $7", transaction.Amount, transaction.Debit, transaction.OffsetAccount, transaction.Account, transaction.Date, transaction.Description, transaction.ID)
	if err != nil {
		return err
	}
	return nil

}

func DeleteTransaction(database *sql.DB, id int) error {
	_, err := database.Exec("DELETE FROM transactions WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil

}

func GetTransaction(database *sql.DB, id int) (Transaction, error) {
	var transaction Transaction
	err := database.QueryRow("SELECT * FROM transactions WHERE id = $1", id).Scan(&transaction.ID, &transaction.Amount, &transaction.Debit, &transaction.OffsetAccount, &transaction.Account, &transaction.Date, &transaction.Description)
	if err != nil {
		return transaction, err
	}
	return transaction, nil
}

func GetTransactions(database *sql.DB, account int, year int, month int) ([]Transaction, error) {
	var transactions []Transaction
	var row *sql.Rows
	var err error
	if year == 0 {
		return nil, errors.New("year is required")
	}

	// TODO: Extract is probably not used right
	if month == 0 {
		row, err = database.Query("SELECT * FROM transactions WHERE account = $1 AND EXTRACT(YEAR FROM time) = $2", account, year)
	} else {
		row, err = database.Query("SELECT * FROM transactions WHERE account = $1 AND EXTRACT(YEAR FROM time) = $2 AND EXTRACT(MONTH FROM time) = $3", account, year, month)
	}
	if err != nil {
		return nil, err
	}

	defer row.Close()

	for row.Next() {
		var transaction Transaction
		err := row.Scan(&transaction.ID, &transaction.Amount, &transaction.Debit, &transaction.OffsetAccount, &transaction.Account, &transaction.Date, &transaction.Description)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil

}

func NewUser(database *sql.DB, user User) error {
	_, err := database.Exec("INSERT INTO users (name, password) VALUES ($1, $2)", user.Name, user.Password)
	if err != nil {
		return err
	}
	return nil

}

func UpdateUser(database *sql.DB, user User) error {
	_, err := database.Exec("UPDATE users SET name = $1, password = $2 WHERE id = $3", user.Name, user.Password, user.ID)
	if err != nil {
		return err
	}
	return nil

}

func DeleteUser(database *sql.DB, id int) error {
	_, err := database.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil

}

func GetUser(database *sql.DB, id int) (User, error) {
	var user User
	err := database.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		return user, err
	}
	return user, nil
}

func GetUserByName(database *sql.DB, name string) (User, error) {
	var user User
	err := database.QueryRow("SELECT * FROM users WHERE name = $1", name).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		return user, err
	}
	return user, nil

}

func AuthenicateUser(database *sql.DB, name string, password string) (User, error) {
	var user User
	err := database.QueryRow("SELECT * FROM users WHERE name = $1 AND password = $2", name, password).Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		return user, err
	}
	return user, nil

}
