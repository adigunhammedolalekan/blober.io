package models

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
	"time"
)

var userRegexp = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
var hostRegexp = regexp.MustCompile("^[^\\s]+\\.[^\\s]+$")

// ErrInvalidEmail is returned when a user's
// email address is invalid
var ErrInvalidEmail = errors.New("Invalid email address")

// ErrInvalidPassword is returned when a user
// use a weak password. less than 5 in chars
var ErrInvalidPassword = errors.New("Invalid password. too weak")

// Account rep the user's/developer's account
type Account struct {
	gorm.Model
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Password string `json:"password"`

	Cred *Credential `json:"cred" gorm:"-" sql:"-"`
}

// NewAccount creates a new account pointer
func NewAccount(firstName, lastName, email, password string) *Account {
	return &Account{
		FirstName: firstName, LastName: lastName, Email: email, Password: password,
		Cred: NewCredential(),
	}
}

// Validate validates important accounts data
func (a *Account) Validate() error {
	if err := validateEmail(a.Email); err != nil {
		return err
	}

	if len(a.Password) < 5 {
		return ErrInvalidPassword
	}

	return nil
}

// Credential is an account credential
type Credential struct {
	PrivateAccessKey string `json:"private_access_key"`
	PublicAccessKey string `json:"public_access_key"`
	ExpiresIn time.Time `json:"expires_in"`
}

// NewCredential generates a unique, random private
// and public key
// critical app features like creating app and fetching lists
// of created apps requires a private key, while uploading files
// requires either private or public key
// public key is useful when uploading files directly from clients(Js, Android, iOS etc)
func NewCredential() *Credential {
	prefix := randomMD5()[:16]

	private := fmt.Sprintf("priv%s%s", prefix, randomSHA256())
	public := fmt.Sprintf("pub%s%s", prefix, randomSHA256())
	return &Credential{PrivateAccessKey: private, PublicAccessKey: public,
		ExpiresIn: time.Now().Add(time.Hour * 25 * 5 /*5 days*/)}
}

// StripKey returns the database key used
// to stored this credential
func (c *Credential) StripKey() string {
	if c.PrivateAccessKey != "" {
		return c.PrivateAccessKey[:20]
	}

	return c.PublicAccessKey[:20]
}
// Username returns a unique identifier for
// this account. (username+Id)
func (a *Account) Username() string {
	parts := strings.Split(a.Email, "@")
	return fmt.Sprintf("%s%d", parts[0], a.ID)
}
// EmailUsername returns username part of
// an email address
func (a *Account) EmailUsername() string {
	return strings.Split(a.Email, "@")[0]
}

// String returns json marshalled string
// of this account
func (a *Account) String() string {
	d, err := json.Marshal(a)
	if err != nil {
		return ""
	}

	return string(d)
}


func randomSHA256() string {
	s := uuid.New().String()
	sh := sha256.New()
	sh.Write([]byte(s))
	sh.Write([]byte(string(time.Now().UnixNano())))
	return fmt.Sprintf("%x", sh.Sum(nil))
}

func randomMD5() string {
	s := uuid.New().String()
	m5 := md5.New()
	m5.Write([]byte(s))
	return fmt.Sprintf("%x", m5.Sum(nil))
}

// validateEmail validates an email address
func validateEmail(email string) error {

	if len(email) < 6 || len(email) > 254 {
		return ErrInvalidEmail
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return ErrInvalidEmail
	}

	user := email[:at]
	host := email[at+1:]

	if len(user) > 64 {
		return ErrInvalidEmail
	}

	if !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return ErrInvalidEmail
	}

	return nil
}
