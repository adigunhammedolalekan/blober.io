package store

import (
	"blober.io/models"
	"github.com/joho/godotenv"
	"log"
	"testing"
	"time"
)

var store *SessionStore
var key = "lekan"

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal(err)
	}
	store, _ = NewSessionStore()
}

var (
	account = &models.Account{FirstName: "Lekan", LastName: "Adigun"}
	session = &Session{
		ID: "1", Created: time.Now(), Account: account,
	}
)

func TestSessionStore_Set(t *testing.T) {
	err := store.Set(key, account)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSessionStore_Get(t *testing.T) {
	a, err := store.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if a.FirstName != account.FirstName {
		t.Fatalf("expected %s, %s returned", account.FirstName, a.FirstName)
	}

	if a.LastName != account.LastName {
		t.Fatalf("expected %s, %s returned", account.LastName, a.LastName)
	}
}
