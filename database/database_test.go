package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("pgx", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	// setup database

	sqlFile, err := os.ReadFile("../database/bookholder.sql")
	if err != nil {
		log.Fatalf("Could not read sql file: %s", err)
	}

	sqlStatements := strings.Split(string(sqlFile), ";")
	for _, statement := range sqlStatements {
		_, err = db.Exec(statement)
		if err != nil {
			log.Fatalf("Could not create tables: %s", err)
		}
	}

	// load sample data
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES (1, 'Test Account', '1000.00')")
	if err != nil {
		log.Fatalf("Could not insert sample data: %s", err)
	}

	// run tests
	m.Run()
}

func cleanTables() {
	_, err := db.Exec("DELETE FROM transactions")
	if err != nil {
		log.Fatalf("Could not clean tables: %s", err)
	}
	_, err = db.Exec("DELETE FROM accounts")
	if err != nil {
		log.Fatalf("Could not clean tables: %s", err)
	}

	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		log.Fatalf("Could not clean tables: %s", err)
	}

}

func TestNewAccount(t *testing.T) {
	cleanTables()
	account := Account{ID: 1000, Name: "Test Account 2", Kind: "1000.00"}
	err := NewAccount(db, account)
	if err != nil {
		t.Error(err)
	}

	// check if account was created
	var id uint
	err = db.QueryRow("SELECT id FROM accounts WHERE id = $1", account.ID).Scan(&id)
	if err != nil {
		t.Error(err)
	}
	t.Log("Retrieved id:", id)
	t.Log("Expected id:", account.ID)
	assert.Equal(t, account.ID, id)
}

func TestNewAccountAlreadyExists(t *testing.T) {
	cleanTables()
	iniAccount := Account{ID: 1, Name: "Test Account", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", iniAccount.ID, iniAccount.Name, iniAccount.Kind)
	if err != nil {
		t.Error(err)
	}

	account := Account{ID: 1, Name: "Test Account 2", Kind: "1000.00"}
	err = NewAccount(db, account)
	if err == nil {
		t.Error("expected error")
	}

	t.Log("Error:", err)
	t.Log("Expected error: account already exists")
	assert.Equal(t, "account already exists", err.Error())
}

func TestExistAccount(t *testing.T) {
	exists, err := existAccount(db, 1)
	if err != nil {
		t.Error(err)
	}

	t.Log("Account exists:", exists)
	t.Log("Expected account exists: true")
	assert.True(t, exists)
}

func TestExistAccountNotExists(t *testing.T) {
	exists, err := existAccount(db, 2)
	if err != nil {
		t.Error(err)
	}

	t.Log("Account exists:", exists)
	t.Log("Expected account exists: false")
	assert.False(t, exists)
}

func TestUpdateAccount(t *testing.T) {
	cleanTables()
	iniAccount := Account{ID: 3, Name: "Test Account", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", iniAccount.ID, iniAccount.Name, iniAccount.Kind)
	if err != nil {
		t.Error(err)
	}

	account := Account{ID: 3, Name: "Test Account 3", Kind: "1000.00"}
	err = UpdateAccount(db, account)
	if err != nil {
		t.Error(err)
	}

	// check if account was updated
	var name string
	err = db.QueryRow("SELECT name FROM accounts WHERE id = $1", account.ID).Scan(&name)
	if err != nil {
		t.Error(err)
	}
	t.Log("Retrieved name:", name)
	t.Log("Expected name:", account.Name)
	assert.Equal(t, account.Name, name)
}

func TestUpdateAccountNotExists(t *testing.T) {
	account := Account{ID: 4, Name: "Test Account 4", Kind: "1000.00"}
	err := UpdateAccount(db, account)
	if err == nil {
		t.Error("expected error")
	}

	t.Log("Error:", err)
	t.Log("Expected error: account does not exist")
	assert.Equal(t, "account does not exist", err.Error())
}

func TestDeleteAccount(t *testing.T) {
	cleanTables()
	account := Account{ID: 5, Name: "Test Account 5", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account.ID, account.Name, account.Kind)
	if err != nil {
		t.Error(err)
	}

	err = DeleteAccount(db, int(account.ID))
	if err != nil {
		t.Error(err)
	}

	// check if account was deleted
	var id uint
	err = db.QueryRow("SELECT id FROM accounts WHERE id = $1", account.ID).Scan(&id)
	if err == nil {
		t.Error("account was not deleted")
	}
}

func TestDeleteAccountNotExists(t *testing.T) {
	err := DeleteAccount(db, 6)
	if err == nil {
		t.Error("expected error")
	}

	t.Log("Error:", err)
	t.Log("Expected error: account does not exist")
	assert.Equal(t, "account does not exist", err.Error())
}

func TestGetAccount(t *testing.T) {
	account := Account{ID: 7, Name: "Test Account 7", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account.ID, account.Name, account.Kind)
	if err != nil {
		t.Error(err)
	}

	acc, err := GetAccount(db, int(account.ID))
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved account:", acc)
	t.Log("Expected account:", account)
	assert.Equal(t, account, acc)
}

func TestGetAccountNotExists(t *testing.T) {
	_, err := GetAccount(db, 8)
	if err == nil {
		t.Error("expected error")
	}

	t.Log("Error:", err)
	t.Log("Expected error: account does not exist")
	assert.Equal(t, "account does not exist", err.Error())
}

func TestExistTransaction(t *testing.T) {
	account1 := Account{ID: 9, Name: "Test Account 9", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 10, Name: "Test Account 10", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 0, Amount: 1234.56, Debit: true, OffsetAccount: 9, Account: 10, Date: time.Now()}
	_, err = db.Exec("INSERT INTO transactions (amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction.Amount, transaction.Debit, transaction.OffsetAccount, transaction.Account, transaction.Date, transaction.Description)
	if err != nil {
		t.Error(err)
	}

	exists, err := existTransaction(db, 1)
	if err != nil {
		t.Error(err)
	}

	t.Log("Transaction exists:", exists)
	t.Log("Expected transaction exists: true")
	assert.True(t, exists)
}

func TestExistTransactionNotExists(t *testing.T) {
	exists, err := existTransaction(db, 2)
	if err != nil {
		t.Error(err)
	}

	t.Log("Transaction exists:", exists)
	t.Log("Expected transaction exists: false")
	assert.False(t, exists)
}

func TestNewTransaction(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 11, Name: "Test Account 11", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 12, Name: "Test Account 12", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 1, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction"}
	err = NewTransaction(db, transaction)
	if err != nil {
		t.Error(err)
	}

	var id uint
	err = db.QueryRow("SELECT id FROM transactions limit 1").Scan(&id)
	if err != nil {
		t.Error(err)
	}

	// check if transaction was created
	var account uint
	err = db.QueryRow("SELECT account FROM transactions WHERE id = $1", id).Scan(&account)
	if err != nil {
		t.Error(err)
	}
	t.Log("Retrieved id:", id)
	t.Log("Expected id:", transaction.ID)
	assert.Equal(t, transaction.Account, account)
}

func TestNewTransactionAccountNotExists(t *testing.T) {
	transaction := Transaction{ID: 2, Amount: 1234.56, Debit: true, OffsetAccount: 13, Account: 14, Date: time.Now(), Description: "Test Transaction"}
	err := NewTransaction(db, transaction)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestNewTransactionOffsetAccountNotExists(t *testing.T) {
	account := Account{ID: 15, Name: "Test Account 15", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account.ID, account.Name, account.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 3, Amount: 1234.56, Debit: true, OffsetAccount: 16, Account: account.ID, Date: time.Now(), Description: "Test Transaction"}
	err = NewTransaction(db, transaction)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestNewTransactionOffsetAccountEqualsAccount(t *testing.T) {
	account := Account{ID: 17, Name: "Test Account 17", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account.ID, account.Name, account.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 4, Amount: 1234.56, Debit: true, OffsetAccount: account.ID, Account: account.ID, Date: time.Now(), Description: "Test Transaction"}
	err = NewTransaction(db, transaction)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestNewTransactionAmountZero(t *testing.T) {
	account1 := Account{ID: 18, Name: "Test Account 18", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 19, Name: "Test Account 19", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 5, Amount: 0, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction"}
	err = NewTransaction(db, transaction)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestNewTransactionDateZero(t *testing.T) {
	account1 := Account{ID: 20, Name: "Test Account 20", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 21, Name: "Test Account 21", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 6, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Time{}, Description: "Test Transaction"}
	err = NewTransaction(db, transaction)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestUpdateTransaction(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 22, Name: "Test Account 22", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 23, Name: "Test Account 23", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	iniTransaction := Transaction{ID: 7, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction"}
	_, err = db.Exec("INSERT INTO transactions (id, amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6, $7)", iniTransaction.ID, iniTransaction.Amount, iniTransaction.Debit, iniTransaction.OffsetAccount, iniTransaction.Account, iniTransaction.Date, iniTransaction.Description)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 7, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction 2"}
	err = UpdateTransaction(db, transaction)
	if err != nil {
		t.Error(err)
	}

	// check if transaction was updated
	var description string
	err = db.QueryRow("SELECT description FROM transactions WHERE id = $1", transaction.ID).Scan(&description)
	if err != nil {
		t.Error(err)
	}
	t.Log("Retrieved description:", description)
	t.Log("Expected description:", transaction.Description)
	assert.Equal(t, transaction.Description, description)
}

func TestUpdateTransactionNotExists(t *testing.T) {
	account1 := Account{ID: 24, Name: "Test Account 24", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 25, Name: "Test Account 25", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 8, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction"}
	err = UpdateTransaction(db, transaction)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestDeleteTransaction(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 26, Name: "Test Account 26", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 27, Name: "Test Account 27", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 9, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction"}
	_, err = db.Exec("INSERT INTO transactions (id, amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6, $7)", transaction.ID, transaction.Amount, transaction.Debit, transaction.OffsetAccount, transaction.Account, transaction.Date, transaction.Description)
	if err != nil {
		t.Error(err)
	}

	err = DeleteTransaction(db, int(transaction.ID))
	if err != nil {
		t.Error(err)
	}

	// check if transaction was deleted
	var id uint
	err = db.QueryRow("SELECT id FROM transactions WHERE id = $1", transaction.ID).Scan(&id)
	if err == nil {
		t.Error("transaction was not deleted")
	}
}

func TestDeleteTransactionNotExists(t *testing.T) {
	err := DeleteTransaction(db, 10)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestGetTransaction(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 28, Name: "Test Account 28", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 29, Name: "Test Account 29", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	transaction := Transaction{ID: 1, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: time.Now(), Description: "Test Transaction"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction.Amount, transaction.Debit, transaction.OffsetAccount, transaction.Account, transaction.Date, transaction.Description)
	if err != nil {
		t.Error(err)
	}

	var id int
	err = db.QueryRow("SELECT id FROM transactions limit 1").Scan(&id)
	if err != nil {
		t.Error(err)
	}

	result, err := GetTransaction(db, id)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved transaction:", result.Amount)
	t.Log("Expected transaction:", transaction.Amount)
	assert.Equal(t, result.Amount, transaction.Amount)
}

func TestGetTransactionNotExists(t *testing.T) {
	_, err := GetTransaction(db, 2)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestGetTransactionsAccountYear(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 30, Name: "Test Account 30", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 31, Name: "Test Account 31", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	date, _ := time.Parse("2006-01-02", "2024-01-01")

	transaction1 := Transaction{ID: 1, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 1"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction1.Amount, transaction1.Debit, transaction1.OffsetAccount, transaction1.Account, transaction1.Date, transaction1.Description)
	if err != nil {
		t.Error(err)
	}

	transaction2 := Transaction{ID: 2, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 2"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction2.Amount, transaction2.Debit, transaction2.OffsetAccount, transaction2.Account, transaction2.Date, transaction2.Description)
	if err != nil {
		t.Error(err)
	}

	transactions, err := GetTransactions(db, 31, 2024, 0)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved transactions:", transactions)
	t.Log("Expected transactions:", []Transaction{transaction1, transaction2})
	assert.Equal(t, transaction1.Amount, transactions[0].Amount)
	assert.Equal(t, transaction2.Amount, transactions[1].Amount)
}

func TestGetTransactionsOfffsetAccountYear(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 30, Name: "Test Account 30", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 31, Name: "Test Account 31", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	date, _ := time.Parse("2006-01-02", "2024-01-01")

	transaction1 := Transaction{ID: 1, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 1"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction1.Amount, transaction1.Debit, transaction1.OffsetAccount, transaction1.Account, transaction1.Date, transaction1.Description)
	if err != nil {
		t.Error(err)
	}

	transaction2 := Transaction{ID: 2, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 2"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction2.Amount, transaction2.Debit, transaction2.OffsetAccount, transaction2.Account, transaction2.Date, transaction2.Description)
	if err != nil {
		t.Error(err)
	}

	transactions, err := GetTransactions(db, 30, 2024, 0)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved transactions:", transactions)
	t.Log("Expected transactions:", []Transaction{transaction1, transaction2})
	assert.Equal(t, transaction1.Amount, transactions[0].Amount)
	assert.Equal(t, transaction2.Amount, transactions[1].Amount)
}

func TestGetTransactionsAccountYearMonth(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 30, Name: "Test Account 30", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 31, Name: "Test Account 31", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	date, _ := time.Parse("2006-01-02", "2024-01-01")

	transaction1 := Transaction{ID: 1, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 1"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction1.Amount, transaction1.Debit, transaction1.OffsetAccount, transaction1.Account, transaction1.Date, transaction1.Description)
	if err != nil {
		t.Error(err)
	}

	transaction2 := Transaction{ID: 2, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 2"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction2.Amount, transaction2.Debit, transaction2.OffsetAccount, transaction2.Account, transaction2.Date, transaction2.Description)
	if err != nil {
		t.Error(err)
	}

	transactions, err := GetTransactions(db, 31, 2024, 1)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved transactions:", transactions)
	t.Log("Expected transactions:", []Transaction{transaction1, transaction2})
	assert.Equal(t, transaction1.Amount, transactions[0].Amount)
	assert.Equal(t, transaction2.Amount, transactions[1].Amount)
}

func TestGetTransactionsOfffsetAccountYearMonth(t *testing.T) {
	cleanTables()
	account1 := Account{ID: 30, Name: "Test Account 30", Kind: "1000.00"}
	_, err := db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account1.ID, account1.Name, account1.Kind)
	if err != nil {
		t.Error(err)
	}

	account2 := Account{ID: 31, Name: "Test Account 31", Kind: "1000.00"}
	_, err = db.Exec("INSERT INTO accounts (id, name, kind) VALUES ($1, $2, $3)", account2.ID, account2.Name, account2.Kind)
	if err != nil {
		t.Error(err)
	}

	date, _ := time.Parse("2006-01-02", "2024-01-01")

	transaction1 := Transaction{ID: 1, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 1"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction1.Amount, transaction1.Debit, transaction1.OffsetAccount, transaction1.Account, transaction1.Date, transaction1.Description)
	if err != nil {
		t.Error(err)
	}

	transaction2 := Transaction{ID: 2, Amount: 1234.56, Debit: true, OffsetAccount: account1.ID, Account: account2.ID, Date: date, Description: "Test Transaction 2"}
	_, err = db.Exec("INSERT INTO transactions ( amount, debit, offset_account, account, date, description) VALUES ($1, $2, $3, $4, $5, $6)", transaction2.Amount, transaction2.Debit, transaction2.OffsetAccount, transaction2.Account, transaction2.Date, transaction2.Description)
	if err != nil {
		t.Error(err)
	}

	transactions, err := GetTransactions(db, 30, 2024, 1)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved transactions:", transactions)
	t.Log("Expected transactions:", []Transaction{transaction1, transaction2})
	assert.Equal(t, transaction1.Amount, transactions[0].Amount)
	assert.Equal(t, transaction2.Amount, transactions[1].Amount)
}

func TestNewUser(t *testing.T) {
	cleanTables()
	user := User{ID: "1", Name: "Test User", Password: "password"}
	err := NewUser(db, user)
	if err != nil {
		t.Error(err)
	}

	// check if user was created
	var Name string
	err = db.QueryRow("SELECT name FROM users WHERE name = $1", user.Name).Scan(&Name)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved id:", Name)
	t.Log("Expected id:", user.Name)
	assert.Equal(t, user.Name, Name)
}

func TestNewUserAlreadyExists(t *testing.T) {
	cleanTables()
	user1 := User{ID: "", Name: "Test User", Password: "password"}
	err := NewUser(db, user1)
	if err != nil {
		t.Error(err)
	}
	user2 := User{ID: "", Name: "Test User", Password: "password"}
	err = NewUser(db, user2)
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestUpdateUser(t *testing.T) {
	cleanTables()
	iniUser := User{ID: "2", Name: "Test User", Password: "password"}
	_, err := db.Exec("INSERT INTO users ( name, password) VALUES ($1, $2)", iniUser.Name, iniUser.Password)
	if err != nil {
		t.Error(err)
	}

	var id string
	err = db.QueryRow("SELECT id FROM users WHERE name = $1", iniUser.Name).Scan(&id)
	if err != nil {
		t.Error(err)
	}

	user := User{ID: id, Name: "Test User 2", Password: "password"}
	err = UpdateUser(db, user)
	if err != nil {
		t.Error(err)
	}

	// check if user was updated
	var name string
	err = db.QueryRow("SELECT name FROM users WHERE name = $1", user.Name).Scan(&name)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved name:", name)
	t.Log("Expected name:", user.Name)
	assert.Equal(t, user.Name, name)
}

func TestDeleteUser(t *testing.T) {
	cleanTables()
	user := User{ID: "3", Name: "Test User", Password: "password"}
	_, err := db.Exec("INSERT INTO users ( name, password) VALUES ($1, $2)", user.Name, user.Password)
	if err != nil {
		t.Error(err)
	}

	var id string
	err = db.QueryRow("SELECT id FROM users WHERE name = $1", user.Name).Scan(&id)
	if err != nil {
		t.Error(err)
	}

	err = DeleteUser(db, id)
	if err != nil {
		t.Error(err)
	}

	// check if user was deleted
	err = db.QueryRow("SELECT id FROM users WHERE id = $1", user.ID).Scan(&id)
	if err == nil {
		t.Error("user was not deleted")
	}
}

func TestGetUser(t *testing.T) {
	cleanTables()
	user := User{ID: "4", Name: "Test User", Password: "password"}
	_, err := db.Exec("INSERT INTO users ( name, password) VALUES ($1, $2)", user.Name, user.Password)
	if err != nil {
		t.Error(err)
	}

	var id string
	err = db.QueryRow("SELECT id FROM users WHERE name = $1", user.Name).Scan(&id)
	if err != nil {
		t.Error(err)
	}

	result, err := GetUser(db, id)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved user:", result.Name)
	t.Log("Expected user:", user.Name)
	assert.Equal(t, user.Name, result.Name)
}

func TestGetUserNotExists(t *testing.T) {
	_, err := GetUser(db, "5")
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestGetUserByName(t *testing.T) {
	cleanTables()
	user := User{ID: "6", Name: "Test User", Password: "password"}
	_, err := db.Exec("INSERT INTO users ( name, password) VALUES ($1, $2)", user.Name, user.Password)
	if err != nil {
		t.Error(err)
	}

	result, err := GetUserByName(db, user.Name)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved user:", result.Name)
	t.Log("Expected user:", user.Name)
	assert.Equal(t, user.Name, result.Name)
}

func TestGetUserByNameNotExists(t *testing.T) {
	_, err := GetUserByName(db, "SHITTY USER NAME")
	if err == nil {
		t.Error("expected error")
	}

	assert.Error(t, err)
}

func TestAuthenticateUser(t *testing.T) {
	cleanTables()
	user := User{ID: "7", Name: "Test User", Password: "password"}
	_, err := db.Exec("INSERT INTO users ( name, password) VALUES ($1, $2)", user.Name, user.Password)
	if err != nil {
		t.Error(err)
	}

	result, err := AuthenticateUser(db, user.Name, user.Password)
	if err != nil {
		t.Error(err)
	}

	t.Log("Retrieved user:", result.Name)
	t.Log("Expected user:", user.Name)
	assert.Equal(t, user.Name, result.Name)
}
