package models

import "testing"

func TestNewAccount(t *testing.T) {
	account := NewAccount("lekan", "hammed", "adigun@gmail.com", "manman")
	if account.FirstName != "lekan" {
		t.Fail()
	}
	if account.LastName != "hammed" {
		t.Fail()
	}
	if account.Email != "adigun@gmail.com" {
		t.Fail()
	}
	if account.Password != "manman" {
		t.Fail()
	}

	if account.Cred.PrivateAccessKey == "" || account.Cred.PublicAccessKey == "" {
		t.Fail()
	}
}

func TestCredential_StripKey(t *testing.T) {
	account := NewAccount("lekan", "hammed", "adigun@gmail.com", "manman")
	key := account.Cred.StripKey()
	if len(key) != 20 {
		t.Fail()
	}
}

func TestAccount_EmailUsername(t *testing.T) {
	account := NewAccount("", "", "adigun@gmail.com", "manman")
	if account.EmailUsername() != "adigun" {
		t.Fatalf("expected %s, found %s", "adigun", account.EmailUsername())
	}
}

func TestAccount_Validate(t *testing.T) {
	account := NewAccount("lekan", "hammed", "adigun@gmail.com", "manman")
	err := account.Validate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccount_ValidateFail(t *testing.T) {
	account := NewAccount("lekan", "hammed", "adigungmail.com", "mman")
	err := account.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with error")
	}
}
