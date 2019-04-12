package store

import (
	"blober.io/models"
	"encoding/json"
	"github.com/dgraph-io/badger"
	"github.com/google/uuid"
	"log"
	"os"
	"time"
)

// sessionExpiration is the total time a session can live
// session would be automatically removed/delete after this time
var sessionExpiration = 2 * 24 * time.Hour // two days

// cleanUpInterval is the time interval at which
// sessionstore would be clean up
var cleanUpInterval = 20 * time.Minute

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

// Session holds info about session
type Session struct {
	ID string `json:"id"`
	Created time.Time `json:"created"`
	Account *models.Account `json:"account"`
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

	sess := &SessionStore{db: db}
	go sess.cleanUp()
	return sess, nil
}

// Set creates and store a new session
func (s *SessionStore) Set(key string, account *models.Account) error {
	session := &Session{ID: uuid.New().String(), Created: time.Now(), Account: account}
	return s.db.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(session)
		if err != nil {
			log.Printf("failed to marshall session %v", err)
			return err
		}

		return txn.Set([]byte(key), b)
	})
}

// Get retrieve a stored session
func (s *SessionStore) Get(key string) (models.Account, error) {
	var session Session
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

		return json.Unmarshal(value, &session)
	})

	return *session.Account, err
}

// cleanUp do a sane, session cleanup.
// Which means expired session would be removed
// and users would have to reauthenticate
func (s *SessionStore) cleanUp() {
	ticker := time.NewTicker(cleanUpInterval)
	for {
		select {
		case <-ticker.C:
			err := s.doCleanUp()
			if err != nil {
				log.Printf("failed to do cleanup %v", err)
			}
		}
	}
}

// doCleanUp deletes all expired sessions
func (s *SessionStore) doCleanUp() error {
	return s.db.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchSize = 20
		it := txn.NewIterator(opt)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			bytes, err := item.Value()
			if err != nil {
				log.Printf("failed to get data value, %v", err)
				continue
			}

			var sess Session
			err = json.Unmarshal(bytes, sess)
			if err != nil {
				log.Printf("failed to get session data %v", err)
			}

			// has the session expired?
			// delete it if it has
			if d := sess.Created.Sub(time.Now()); d > sessionExpiration {
				err = txn.Delete(key)
				if err != nil {
					log.Printf("failed to delete item, key = %s. Reason = %v", key, err)
				}
			}
		}

		return nil
	})
}

// Close closes underling badgerDB
func (s *SessionStore) Close() error {
	return s.db.Close()
}