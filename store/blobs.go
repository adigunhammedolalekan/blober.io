package store

import (
	"blober.io/models"
	"encoding/json"
	"github.com/dgraph-io/badger"
	"log"
	"os"
)

// BlobStore caches Blob/File struct data
// for easy access when being download,
// it should provide faster access and be more
// resources efficient than say a remote SQL database.
// It is back by a key/value store(BadgerDB)
type BlobStore struct {
	db *badger.DB
}

// NewBlobStore creates a new blobstore
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

// Set puts/caches a blob
// key is unique for each blob
// key in most cases equals appname+hash
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

// Get fetches a cached blob
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

// Close underlying badgerDB
func (s *BlobStore) Close() error {
	return s.db.Close()
}