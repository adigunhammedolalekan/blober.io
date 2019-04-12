package store

import (
	"blober.io/models"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

var blobStore *BlobStore
var accessKey = "app1"
var blob = &models.Blob{
	Hash: "9876HJKL", AppName: "appOne",
}

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal(err)
	}
	blobStore, _ = NewBlobStore()
}

func TestBlobStore_Set(t *testing.T) {
	err := blobStore.Set(accessKey, blob)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBlobStore_Get(t *testing.T) {
	b, err := blobStore.Get(accessKey)
	if err != nil {
		t.Fatal(err)
	}

	if b.Hash != blob.Hash {
		t.Fatalf("expected hash %s, found %s", blob.Hash, b.Hash)
	}
}
