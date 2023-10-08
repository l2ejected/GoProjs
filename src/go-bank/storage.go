package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(string) (*Account, error)
	GetAccounts() ([]*Account, error)
	TransferMoney(giver, recepient *Account, amount float64) error
	Debit(*Account, float64) error
	Credit(*Account, float64) error
}

type PostGresStore struct {
	db *sql.DB
}

func NewPostGresStore() (*PostGresStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostGresStore{
		db: db,
	}, nil
}

func (s *PostGresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostGresStore) createAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		uuid char(36),
		balance decimal(19,4),
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostGresStore) CreateAccount(a *Account) error {
	query := `
	insert into account
	(first_name, last_name, uuid, balance, created_at)
	values ($1, $2, $3, $4, $5) 
	`

	res, err := s.db.Exec(query, a.FirstName, a.LastName, a.UUID, a.Balance, a.CreatedAt)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Printf("added %d new account\n", affected)

	return nil
}

func (s *PostGresStore) UpdateAccount(acc *Account) error {
	query := `
		UPDATE account
		SET first_name = $1, last_name = $2, balance = $3
		WHERE id = $4
	`
	res, err := s.db.Exec(query, acc.FirstName, acc.LastName, acc.Balance, acc.ID)
	if err != nil {
		return err
	}
	updated, _ := res.RowsAffected()
	fmt.Printf("rows updated: %d \n", updated)

	return nil
}

func (s *PostGresStore) DeleteAccount(id int) error {
	query := "DELETE FROM account WHERE id = $1"
	res, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	fmt.Printf("rows deleted: %d \n", affected)
	return nil
}

func (s *PostGresStore) GetAccountByID(id string) (*Account, error) {
	query := `
		SELECT * from account
		WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	account := new(Account)
	err := row.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.UUID,
		&account.Balance,
		&account.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("account with id = %s does not exist", id)
		}
		return nil, err
	}

	return account, nil
}

func (s *PostGresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account := new(Account)
		err := rows.Scan(
			&account.ID,
			&account.FirstName,
			&account.LastName,
			&account.UUID,
			&account.Balance,
			&account.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *PostGresStore) TransferMoney(giver, recepient *Account, amount float64) error {
	if err := s.Debit(recepient, amount); err != nil {
		return err
	}

	if err := s.Credit(giver, amount); err != nil {
		return err
	}

	return nil
}

func (s *PostGresStore) Debit(acc *Account, amt float64) error {
	newBalance := acc.Balance + amt
	query := `
		UPDATE account
		SET balance = $1
		WHERE id = $2
	`
	if _, err := s.db.Exec(query, newBalance, acc.ID); err != nil {
		return err
	}
	return nil
}

func (s *PostGresStore) Credit(acc *Account, amt float64) error {
	newBalance := acc.Balance - amt
	query := `
		UPDATE account
		SET balance = $1
		WHERE id = $2
	`
	if _, err := s.db.Exec(query, newBalance, acc.ID); err != nil {
		return err
	}
	return nil
}
