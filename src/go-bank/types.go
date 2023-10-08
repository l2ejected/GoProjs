package main

import (
	"time"

	"github.com/google/uuid"
)

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TransferBalanceRequest struct {
	Recepient string  `json:"id"`
	Amount    float64 `json:"amount"`
}

type BalanceChangeRequest struct {
	Amount float64 `json:"amount"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	UUID      uuid.UUID `json:"uuid"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(req CreateAccountRequest) *Account {
	return &Account{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		UUID:      uuid.New(),
		CreatedAt: time.Now().UTC(),
	}
}

func GetUpdatedAccount(old *Account, toUpdate *Account) *Account {

	if toUpdate.FirstName != "" && old.FirstName != toUpdate.FirstName {
		old.FirstName = toUpdate.FirstName
	}

	if toUpdate.LastName != "" && old.LastName != toUpdate.LastName {
		old.LastName = toUpdate.LastName
	}

	if old.Balance != toUpdate.Balance {
		old.Balance = toUpdate.Balance
	}

	return old
}
