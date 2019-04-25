package models

import "testing"

func TestNewApp(t *testing.T) {
	app := NewApp("appName", 1)
	if app.Name != "appName" {
		t.Fail()
	}

	if app.AccountId != 1 {
		t.Fail()
	}
}

func TestApp_Validate(t *testing.T) {
	app := NewApp("appName", 1)
	err := app.Validate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestApp_ValidateError(t *testing.T) {
	app := NewApp("", 0)
	err := app.Validate()
	if err == nil {
		t.Fatal("Validate() should return error for invalid data")
	}
}

func TestApp_UniqueId(t *testing.T) {
	app := NewApp("appName", 1)
	if app.UniqueId() != "appName" {
		t.Fail()
	}
}

func TestNewBlob(t *testing.T) {

	hash := "09876HGFiuyt"
	blob := NewBlob(hash, "img", &App{Name: "appName", AccountId: 1}, 200)
	if blob.Hash != hash {
		t.Fail()
	}
	if blob.ContentType != "img" {
		t.Fail()
	}
	if blob.Size != 200 {
		t.Fail()
	}
}

func TestBlob_BlobDownloadURL(t *testing.T) {
	hash := "09876HGFiuyt"
	blob := NewBlob(hash, "img", &App{Name: "appName", AccountId: 1}, 200)

	downloadUrl := "http://blober.io/res/appName/09876HGFiuyt"
	if blob.BlobDownloadURL() != downloadUrl {
		t.Fatalf("download URL should be %s, found %s", downloadUrl, blob.BlobDownloadURL())
	}
}

func TestBlob_PopulateDownloadURL(t *testing.T) {
	hash := "09876HGFiuyt"
	blob := NewBlob(hash, "img", &App{Name: "appName", AccountId: 1}, 200)
	blob.PopulateDownloadURL()
	downloadUrl := "http://blober.io/res/appName/09876HGFiuyt"

	if blob.DownloadURL != downloadUrl {
		t.Fatalf("expected downloadURL to be set as %s, found %s", downloadUrl, blob.DownloadURL)
	}
}
