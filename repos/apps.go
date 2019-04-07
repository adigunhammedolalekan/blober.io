package repos

import (
	"blober.io/models"
	"blober.io/services"
	"errors"
	"github.com/jinzhu/gorm"
	"io"
	"log"
	"mime/multipart"
)

var Limit int64 = 20

type AppRepository struct {
	db *gorm.DB
	account *AccountRepository
	storage *services.StorageService
}

func NewAppRepository(db *gorm.DB, account *AccountRepository, storage *services.StorageService) *AppRepository {
	return &AppRepository{db:db, account: account, storage:storage}
}

func (repo *AppRepository) CreateNewApp(account uint, name string) (*models.App, error)  {
	existingApp := repo.GetAppByAttr("name", name)
	if existingApp != nil {
		return nil, errors.New("an app with that name already exists")
	}

	app := models.NewApp(name, account)
	if err := app.Validate(); err != nil {
		return nil, err
	}

	user := repo.account.GetAccountByAttr("id", account)
	if user == nil {
		return nil, errors.New("user account not found")
	}

	tx := repo.db.Begin()
	err := tx.Error
	if err != nil {
		return nil, err
	}

	if err := tx.Create(app).Error; err != nil {
		log.Printf("failed to create app, %v", err)
		tx.Rollback()
		return nil, err
	}

	app.Account = user
	if err := repo.storage.CreateBucketForApp(app); err != nil {
		log.Printf("failed to create bucket %v", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("failed to commit create app transaction, %v", err)
		return nil, err
	}

	return app, nil
}

func (repo *AppRepository) UploadBlob(account uint, appName string, private bool, body *multipart.FileHeader) (*models.Blob, error) {

	app := repo.GetAppByName(account, appName)
	if app == nil {
		return nil, errors.New("app not found")
	}

	file, err := body.Open()
	if err != nil {
		return nil, err
	}

	blob, err := repo.storage.UploadBlob(app, private, file)
	if err != nil {
		return nil, err
	}

	blob.Filename = body.Filename
	if err := repo.db.Create(blob).Error; err != nil {
		return nil, err
	}

	return blob, nil
}

func (repo *AppRepository) GetAccountApps(accountId uint) []*models.App {
	data := make([]*models.App, 0)
	err := repo.db.Table("apps").Where("account_id = ?", accountId).Find(&data).Error
	if err != nil {
		return nil
	}

	return data
}

func (repo *AppRepository) DownloadBlob(appName, hash string) (io.Reader, *models.Blob, error) {
	blob, err := repo.storage.GetBlob(appName, hash)
	if err != nil {
		log.Printf("failed to get blob, %v", err)
		return nil, nil, err
	}

	file, err := repo.storage.GetFile(appName, hash)
	if err != nil {
		log.Printf("failed to get file from minio, %v", err)
		return nil, nil, err
	}

	return file, &blob, nil
}

func (repo *AppRepository) GetAppByName(account uint, appName string) *models.App {
	app := &models.App{}
	err := repo.db.Table("apps").Where("account_id = ? AND name = ?", account, appName).First(app).Error
	if err != nil {
		return nil
	}

	if app.AccountId > 0 {
		app.Account = repo.account.GetAccountByAttr("id", app.AccountId)
	}
	return app
}


func (repo *AppRepository) GetAppByAttr(attr string, value interface{}) *models.App {
	app := &models.App{}
	err := repo.db.Table("apps").Where(attr + " = ?", value).First(app).Error
	if err != nil {
		return nil
	}

	return app
}

func (repo *AppRepository) GetAppBlobs(appId uint, page int64) []*models.Blob {
	data := make([]*models.Blob, 0)
	err := repo.db.Table("blobs").Where("app_id = ?", appId).Offset(page * Limit).Limit(Limit).
		Find(&data).Error
	if err != nil {
		return nil
	}

	return data
}