package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
)

// App is an application created by a blober.io user
type App struct {
	gorm.Model
	Name string `json:"name"`
	AccountId uint `json:"account_id"`

	Account *Account `json:"account" sql:"-" gorm:"-"`
}

// Blob is an uploaded file on blober.io
type Blob struct {
	gorm.Model
	AppId uint `json:"app_id"`
	Hash string `json:"hash"`
	Size int64 `json:"size"`
	ContentType string `json:"content_type"`
	DownloadURL string `json:"download_url"`
	AppName string `json:"app_name"`
	IsPrivate bool `json:"is_private"`

	App *App `json:"-" gorm:"-" sql:"-"`
}

// NewApp creates new app
func NewApp(name string, account uint) *App {
	return &App{Name:name, AccountId: account}
}

// Validate validates app request body
func (a *App) Validate() error {
	if len(a.Name) == 0 {
		return errors.New("app can not have an empty name")
	}

	if a.AccountId <= 0 {
		return errors.New("invalid account")
	}

	return nil
}

// UniqueId returns a unique identifier for this app.
// useful for creating minio buckets
func (a *App) UniqueId() string {
	return fmt.Sprintf("%s", a.Name)
}

// NewBlob create a new blob
func NewBlob(hash, contentType string, app *App, size int64) *Blob {
	b := &Blob{
		Hash: hash, ContentType: contentType, AppId:app.ID, Size:size, App:app,
	}

	b.AppName = app.Name
	b.PopulateDownloadURL()
	return b
}

func (b *Blob) PopulateDownloadURL() {
	b.DownloadURL = fmt.Sprintf("%s%s%s%s", "http://blober.io/", b.App.UniqueId(), "/", b.Hash)
}

// BlobDownloadURL forms a file download url
// useful to download files directly from clients
func (b *Blob) BlobDownloadURL() string {
	return fmt.Sprintf("%s%s%s", "http://blober.io/", b.App.UniqueId(), b.Hash)
}