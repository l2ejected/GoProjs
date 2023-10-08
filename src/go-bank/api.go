package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHTTPHandleFunc(s.handleAccountByID))
	router.HandleFunc("/account/{id}/transfer", makeHTTPHandleFunc(s.handleTransfer))
	router.HandleFunc("/account/{id}/debit", makeHTTPHandleFunc(s.handleDebit))
	router.HandleFunc("/account/{id}/credit", makeHTTPHandleFunc(s.handleCredit))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("%s method not allowed", r.Method)
}

func (s *APIServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccountByID(w, r)
	}
	if r.Method == "PUT" {
		return s.handleUpdateAccount(w, r)
	}

	return fmt.Errorf("%s method not allowed", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(&createAccReq); err != nil {
		return err
	}

	account := NewAccount(*createAccReq)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}
func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	acc := new(Account)
	if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
		return err
	}

	err := s.store.DeleteAccount(acc.ID)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, fmt.Sprintf("removed acc with id = %d", acc.ID))
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	toUpdate := new(Account)
	if err := json.NewDecoder(r.Body).Decode(&toUpdate); err != nil {
		return err
	}

	input := GetUpdatedAccount(account, toUpdate)

	err = s.store.UpdateAccount(input)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, fmt.Sprintf("updated acc with id = %d", account.ID))
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	giver, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	req := new(TransferBalanceRequest)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	recepient, err := s.store.GetAccountByID(req.Recepient)
	if err != nil {
		return err
	}

	err = s.store.TransferMoney(giver, recepient, req.Amount)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK,
		fmt.Sprintf(
			"$%.2f transferred from %s's account with id=%s to %s's account id=%s",
			req.Amount, giver.FirstName, id, recepient.FirstName, req.Recepient,
		),
	)
}

func (s *APIServer) handleDebit(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	req := new(BalanceChangeRequest)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	if err = s.store.Debit(account, req.Amount); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, fmt.Sprintf("$%.2f debited to %s's account with id=%s", req.Amount, account.FirstName, id))
}

func (s *APIServer) handleCredit(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	req := new(BalanceChangeRequest)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	if err = s.store.Credit(account, req.Amount); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, fmt.Sprintf("$%.2f credited from %s's account with id=%s", req.Amount, account.FirstName, id))
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
