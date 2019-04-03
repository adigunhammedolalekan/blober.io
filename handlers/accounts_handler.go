package handlers

import (
	"blober.io/models"
	"blober.io/repos"
	"encoding/json"
	"net/http"
)

type AccountHandler struct {
	repo *repos.AccountRepository
}

func NewAccountHandler(repo *repos.AccountRepository) *AccountHandler {
	return &AccountHandler{repo:repo}
}

func (handler *AccountHandler) CreateNewAccountHandler(w http.ResponseWriter, r *http.Request) {
	// parse JSON body
	payload := &models.Account{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(payload); err != nil {
		BadRequestResponse(w)
		return
	}

	// create account
	newAccount, err := handler.repo.CreateNewAccount(payload.FirstName, payload.LastName,
		payload.Email, payload.Password)
	if err != nil {
		JSON(w, 200, &Response{Error: true, Message: err.Error()})
		return
	}

	newAccount.Password = ""
	JSON(w, 201, &Response{Error: false, Message: "account created", Data: newAccount})
}

func (handler *AccountHandler) AuthenticateAccountHandler(w http.ResponseWriter, r *http.Request) {
	// parse JSON body
	payload := &models.Account{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(payload); err != nil {
		BadRequestResponse(w)
		return
	}

	account, err := handler.repo.AuthenticateAccount(payload.Email, payload.Password)
	if err != nil {
		JSON(w, 200, &Response{Error: true, Message: err.Error()})
		return
	}

	account.Password = ""
	JSON(w, 200, &Response{Error: false, Message: "account authenticated", Data: account})
}
