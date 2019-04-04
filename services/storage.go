package services

import (
	"blober.io/models"
	"blober.io/store"
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"io"
	"log"
	"net/http"
	"strings"
)

type StorageService struct {
	client *minio.Client
	store *store.BlobStore
}

type Option struct {
	AccessKey string
	SecretKey string
	Host string
	Store *store.BlobStore
}

func NewStorageService(opt *Option) (*StorageService, error) {
	client, err := minio.New(opt.Host, opt.AccessKey, opt.SecretKey, false)
	if err != nil {
		return nil, err
	}

	return &StorageService{client:client, store:opt.Store}, nil
}

func (service *StorageService) CreateBucketForApp(app *models.App) error {
	location := "us-east-1"
	bucketName := strings.ToLower(app.UniqueId())
	return service.client.MakeBucket(bucketName, location)
}

func (service *StorageService) UploadBlob(app *models.App, isPrivate bool, body io.Reader) (*models.Blob, error) {
	bucketName := strings.ToLower(app.UniqueId())
	fileName := randomMD5()

	var buf bytes.Buffer
	n, err := io.Copy(&buf, body)
	if err != nil {
		log.Printf("failed to copy file %v", err)
		return nil, err
	}

	contentType := http.DetectContentType(buf.Bytes())
	size, err := service.client.PutObject(bucketName, fileName, body, n,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return nil, err
	}

	blob := models.NewBlob(fileName, contentType, app, size)
	blob.IsPrivate = isPrivate
	key := fmt.Sprintf("%s%s", bucketName, fileName)
	err = service.store.Set(key, blob)
	if err != nil {
		log.Printf("failed to cache blob, %v", err)
		return nil, err
	}

	return blob, nil
}

func (service *StorageService) GetFile(appName, hash string) (io.Reader, error) {
	bucketName := strings.ToLower(appName)
	return service.client.GetObject(bucketName, hash, minio.GetObjectOptions{})
}

func (service *StorageService) GetBlob(appName, hash string) (models.Blob, error) {
	key := fmt.Sprintf("%s%s", strings.ToLower(appName), hash)
	return service.store.Get(key)
}

func randomMD5() string {
	s := uuid.New().String()
	m5 := md5.New()
	m5.Write([]byte(s))
	return fmt.Sprintf("%x", m5.Sum(nil))
}