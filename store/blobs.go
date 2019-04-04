package store

import (
	"blober.io/models"
	"encoding/json"
	"github.com/dgraph-io/badger"
	"log"
	"os"
)

type BlobStore struct {
	db *badger.DB
}

//
func NewBlobStore() (*BlobStore, error) {
	opt := badger.DefaultOptions
	opt.Dir = os.Getenv("DB_DIR") + "/blobs"
	opt.ValueDir = os.Getenv("DB_DIR") + "/blobs"

	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	return &BlobStore{db: db}, nil
}


func (s *BlobStore) Set(key string, blob *models.Blob) error {
	return s.db.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(blob)
		if err != nil {
			log.Printf("failed to marshall blob %v", err)
			return err
		}

		return txn.Set([]byte(key), b)
	})
}

// Get retrieve an user's account details
// mainly to perform authorization check
func (s *BlobStore) Get(key string) (models.Blob, error) {
	var blob models.Blob
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

		return json.Unmarshal(value, &blob)
	})

	return blob, err
}

func (s *BlobStore) Close() error {
	return s.db.Close()
}