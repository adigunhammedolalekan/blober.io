package services

import (
	"blober.io/models"
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
}

type Option struct {
	AccessKey string
	SecretKey string
	Host string
}

func NewStorageService(opt *Option) (*StorageService, error) {
	client, err := minio.New(opt.Host, opt.AccessKey, opt.SecretKey, false)
	if err != nil {
		return nil, err
	}

	return &StorageService{client:client}, nil
}

func (service *StorageService) CreateBucketForApp(app *models.App) error {
	location := "us-east-1"
	bucketName := strings.ToLower(app.UniqueId())
	return service.client.MakeBucket(bucketName, location)
}

func (service *StorageService) UploadBlob(app *models.App, body io.Reader) (*models.Blob, error) {
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

	return models.NewBlob(fileName, app, size), nil
}

func randomMD5() string {
	s := uuid.New().String()
	m5 := md5.New()
	m5.Write([]byte(s))
	return fmt.Sprintf("%x", m5.Sum(nil))
}