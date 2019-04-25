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

// apps query paging limit
var Limit int64 = 20

// AppRepository encapsulates database
// access dealing with apps
type AppRepository struct {
	db      *gorm.DB
	account *AccountRepository
	storage *services.StorageService
}

// NewAppRepository creates new AppRepository
func NewAppRepository(db *gorm.DB, account *AccountRepository, storage *services.StorageService) *AppRepository {
	return &AppRepository{db: db, account: account, storage: storage}
}

// CreateNewApp creates a new app
// app must have a unique name across the platform
func (repo *AppRepository) CreateNewApp(account uint, name string) (*models.App, error) {
	// get existing app with the same name
	existingApp := repo.GetAppByAttr("name", name)
	if existingApp != nil {
		return nil, errors.New("an app with that name already exists")
	}

	// validates supplied name
	app := models.NewApp(name, account)
	if err := app.Validate(); err != nil {
		return nil, err
	}

	// makes sure user account exists in the system
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

// UploadBlob uploads a file
// the file is uploaded to minio server
// using services.StorageService{}
func (repo *AppRepository) UploadBlob(account uint, appName string, private bool, body *multipart.FileHeader) (*models.Blob, error) {

	// make sure app exists
	app := repo.GetAppByName(account, appName)
	if app == nil {
		return nil, errors.New("app not found")
	}

	// converts file to io.Reader
	file, err := body.Open()
	if err != nil {
		return nil, err
	}

	// send file to minio server
	blob, err := repo.storage.UploadBlob(app, private, file)
	if err != nil {
		return nil, err
	}

	// create file record in the database
	blob.Filename = body.Filename
	if err := repo.db.Create(blob).Error; err != nil {
		return nil, err
	}

	return blob, nil
}

// GetAccountApps fetch apps created by accountId
func (repo *AppRepository) GetAccountApps(accountId uint) []*models.App {
	data := make([]*models.App, 0)
	err := repo.db.Table("apps").Where("account_id = ?", accountId).Find(&data).Error
	if err != nil {
		return nil
	}

	return data
}

// DownloadBlob downloads a blob from minio server
// apps are identified by their name and a random md5 hash
func (repo *AppRepository) DownloadBlob(appName, hash string) (io.Reader, *models.Blob, error) {

	// get blob struct from badgerDB
	blob, err := repo.storage.GetBlob(appName, hash)
	if err != nil {
		log.Printf("failed to get blob, %v", err)
		return nil, nil, err
	}

	// get file from minio server
	file, err := repo.storage.GetFile(appName, hash)
	if err != nil {
		log.Printf("failed to get file from minio, %v", err)
		return nil, nil, err
	}

	return file, &blob, nil
}

// GetAppByName get app where name == appName
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

// GetAppByAttr get app where attr == value
func (repo *AppRepository) GetAppByAttr(attr string, value interface{}) *models.App {
	app := &models.App{}
	err := repo.db.Table("apps").Where(attr+" = ?", value).First(app).Error
	if err != nil {
		return nil
	}

	return app
}

// GetAppBlobs returns all files that belongs
// to appId, paginated. A single call returns 20 items
func (repo *AppRepository) GetAppBlobs(appId uint, page int64) []*models.Blob {
	data := make([]*models.Blob, 0)
	err := repo.db.Table("blobs").Where("app_id = ?", appId).Offset(page * Limit).Limit(Limit).
		Find(&data).Error
	if err != nil {
		return nil
	}

	return data
}
