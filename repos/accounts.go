package repos

import (
	"blober.io/models"
	"blober.io/store"
	"errors"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"log"
)

// AccountRepository encapsulates database
// access dealing with user accounts
type AccountRepository struct {
	store *store.SessionStore // session store
	db *gorm.DB               // database handle
}

// NewAccountRepository creates a new repository
func NewAccountRepository(s *store.SessionStore, db *gorm.DB) *AccountRepository {
	return &AccountRepository{store:s, db:db}
}

// CreateNewAccount create a new user/developer account
// it also creates a unique credentials(private and public access key)
// the newly created account are then stored in a sessionstore backed by
// badgerDB, for authentication and authorization for subsequent actions
func (repo *AccountRepository) CreateNewAccount(firstName, lastName, email, password string) (*models.Account, error) {
	existingAccount := repo.GetAccountByAttr("email", email)
	if existingAccount != nil {
		return nil, errors.New("an account is already linked with that email. Please use a different" +
			" email address and retry")
	}

	account := models.NewAccount(firstName, lastName, email, password)
	if err := account.Validate(); err != nil {
		return nil, err
	}

	account.Password = repo.hashPassword(password)
	tx := repo.db.Begin()
	err := tx.Error
	if err != nil {
		return nil, err
	}

	// database interaction is done in a transaction,
	// all OP are rolled back whenever an error
	// is encountered
	if err := tx.Create(account).Error; err != nil {
		log.Printf("failed to create a new account %v", err)
		tx.Rollback()
		return nil, err
	}

	if err := repo.store.Set(account.Cred.StripKey(), account); err != nil {
		log.Printf("failed to create session %v", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("failed to commit account creation db writes %v", err)
		return nil, err
	}

	account.Password = ""
	return account, nil
}

// AuthenticateAccount authenticates an account
// and create a credential then store the account
// in sessionstore
func (repo *AccountRepository) AuthenticateAccount(email, password string) (*models.Account, error) {
	account := repo.GetAccountByAttr("email", email)
	if account == nil {
		return nil, errors.New("invalid email and password combination")
	}

	if ok := repo.comparePassword(account.Password, password); !ok {
		return nil, errors.New("invalid email and password combination")
	}

	account.Cred = models.NewCredential()
	account.Password = ""

	if err := repo.store.Set(account.Cred.StripKey(), account); err != nil {
		log.Printf("failed to set session %v", err)
		return nil, err
	}

	return account, nil
}

// GetAccountByAttr get accounts where attr == value
func (repo *AccountRepository) GetAccountByAttr(attr string, value interface{}) *models.Account {
	account := models.Account{}
	err := repo.db.Table("accounts").Where(attr + " = ?", value).First(&account).Error
	if err != nil {
		return nil
	}

	return &account
}

// hashPassword returns a bcrypted password
func (repo *AccountRepository) hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}

	return string(hashedPassword)
}

// comparePassword compares a bcrypted password
// and a plain one. returns true if matched or false otherwise
func (repo *AccountRepository) comparePassword(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if err != nil {
		return false
	}

	return true
}

