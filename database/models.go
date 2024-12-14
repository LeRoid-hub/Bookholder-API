package database

import "time"

type Account struct {
	ID   uint
	Name string
	Kind string
}

type Transaction struct {
	ID            uint
	Amount        float32
	Debit         bool
	OffsetAccount uint
	Account       uint
	Time          time.Time
	Description   string
}

type User struct {
	ID       uint
	Name     string
	Password string
}
