package store

import (
	"blober.io/models"
	"encoding/json"
	"github.com/dgraph-io/badger"
	"log"
	"os"
)

type SessionStore struct {
	db *badger.DB
}

// NewSessionStore creates a new session store
func NewSessionStore() (*SessionStore, error) {
	opt := badger.DefaultOptions
	opt.Dir = os.Getenv("SESSION_STORE_DIR")
	opt.ValueDir = os.Getenv("SESSION_STORE_DIR")

	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return &SessionStore{db: db}, nil
}

// Set store user account details
// for easy access. For authorization
func (s *SessionStore) Set(key string, account *models.Account) error {
	return s.db.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(account)
		if err != nil {
			log.Printf("failed to marshall account %v", err)
			return err
		}

		return txn.Set([]byte(key), b)
	})
}

// Get retrieve an user's account details
// mainly to perform authorization check
func (s *SessionStore) Get(key string) (models.Account, error) {
	var account models.Account
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			log.Printf("failed to get data for key %s, error %v", key, err)
			return err
		}

		value, err := item.Value()
		if err != nil {
			log.Printf("failed to get item value, key => %s, error => %v", key, err)
			return err
		}

		return json.Unmarshal(value, &account)
	})

	return account, err
}

func (s *SessionStore) Close() error {
	return s.db.Close()
}