package store

import (
	"blober.io/models"
	"encoding/json"
	"github.com/dgraph-io/badger"
	"log"
	"os"
)


// SessionStore caches accounts struct data
// for easy access during authentication,
// because we are using either private or public
// key for authentication, JWT might not be useful.
// it should provide faster access and be more
// resources efficient than say a remote SQL database.
// It is back by a key/value store(BadgerDB)
type SessionStore struct {
	db *badger.DB
}

// NewSessionStore creates a new session store
func NewSessionStore() (*SessionStore, error) {
	opt := badger.DefaultOptions
	opt.Dir = os.Getenv("DB_DIR")
	opt.ValueDir = os.Getenv("DB_DIR")

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

// Close closes underling badgerDB
func (s *SessionStore) Close() error {
	return s.db.Close()
}