package database

type Account struct {
	Number int
	Name   string
	Kind   string
}

type Transaction struct {
	Amount        float32
	Debit         bool
	OffsetAccount int
	Account       int
	Description   string
}

type User struct {
	Name     string
	Password string
}
