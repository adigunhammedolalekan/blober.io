package handlers

import (
	"blober.io/models"
	"blober.io/repos"
	"blober.io/store"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
)

// max memory for multipart form data
var maxMemory int64 = 1024 << 20

// AppHandler handles all app related
// http requests
type AppHandler struct {
	store *store.SessionStore
	blobStore *store.BlobStore
	repo *repos.AppRepository
}

// NewAppHandler creates a new AppHandler
func NewAppHandler(store *store.SessionStore, blobStore *store.BlobStore, repo *repos.AppRepository) *AppHandler {
	return &AppHandler{repo:repo, blobStore:blobStore, store:store}
}

// CreateNewAppHandler handles requests to create a new app
func (handler *AppHandler) CreateNewAppHandler(w http.ResponseWriter, r *http.Request) {

	// get auth key or errored out
	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}
	// sessions are mapped to first 20chars
	// of their private/public key
	// it is unique to all users
	sessionKey := key[:20]
	account, err := handler.store.Get(sessionKey)
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	// check for credential validity
	cred := account.Cred
	if cred.PrivateAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	// parse request body and create the app
	app := &models.App{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(app); err != nil {
		BadRequestResponse(w)
		return
	}

	newApp, err := handler.repo.CreateNewApp(account.ID, app.Name)
	if err != nil {
		JSON(w, 200, &Response{Error: true, Message: err.Error()})
		return
	}

	JSON(w, 201, &Response{Error: false, Message: "app created", Data: newApp})
}
// GetAccountAppsHandler handles request to
// fetch authenticated account's created apps
func (handler *AppHandler) GetAccountAppsHandler(w http.ResponseWriter, r *http.Request)  {
	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}

	account, err := handler.store.Get(key[:20])
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	cred := account.Cred
	if cred.PrivateAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	data := handler.repo.GetAccountApps(account.ID)
	JSON(w, 200, &Response{Error: false, Message: "success", Data: data})
}

// UploadBlobHandler handles file upload requests
func (handler *AppHandler) UploadBlobHandler(w http.ResponseWriter, r *http.Request) {

	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}

	account, err := handler.store.Get(key[:20])
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	// files can be uploaded with either private or public key
	cred := account.Cred
	if cred.PrivateAccessKey != key && cred.PublicAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	file, header, err := r.FormFile("file_data")
	if err != nil {
		BadRequestResponse(w)
		return
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close uploaded file %v", err)
		}
	}()

	// get the name of the app
	// that owns the about to be saved file,
	// returns a BadRequest response if this name
	// is not present
	vars := mux.Vars(r)
	appName := vars["appName"]
	if appName == "" {
		BadRequestResponse(w)
		return
	}

	// make the uploaded file private if specified
	// by client.
	// private files can only be downloaded
	// with private keys
	private := r.FormValue("private")
	isPrivate := private == "true"

	// upload file.
	blob, err := handler.repo.UploadBlob(account.ID, appName, isPrivate, header)
	if err != nil {
		JSON(w, 200, &Response{Error: true, Message: err.Error()})
		return
	}

	JSON(w, 200, &Response{Error: false, Message: "success", Data: blob})
}

// UploadMultipleBlobsHandler handles requests to upload
// multiple files at once
func (handler *AppHandler) UploadMultipleBlobsHandler(w http.ResponseWriter, r *http.Request) {
	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}

	account, err := handler.store.Get(key[:20])
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	// can use either private or public key
	// to upload files
	cred := account.Cred
	if cred.PrivateAccessKey != key && cred.PublicAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	// name of the app that owns
	// the about to be uploaded file
	vars := mux.Vars(r)
	appName := vars["appName"]
	if appName == "" {
		BadRequestResponse(w)
		return
	}

	// parse it! The uploaded files
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		BadRequestResponse(w)
		return
	}

	errorCount := 0 // how many uploads failed?
	successCount := 0 // how many uploads succeeded?

	// slice to store successfully uploaded
	// blobs
	blobs := make([]*models.Blob, 0)
	files := r.MultipartForm.File["files[]"]

	// returns bad request in case of
	// no file attached
	if len(files) == 0 {
		JSON(w, 400, &Response{Error: true, Message: "no file found"})
		return
	}

	// should the files be private?
	isPrivate := r.FormValue("private") == "true"
	for _, value := range files {

		// do the upload
		blob, err := handler.repo.UploadBlob(account.ID, appName, isPrivate, value)
		if err != nil {
			log.Printf("failed to process upload, %v", err)
			errorCount += 1
			continue
		}

		blobs = append(blobs, blob)
		successCount += 1
	}

	// put together the response body and respond
	response := &models.UploadMultipleResponse{SuccessCount: int64(successCount),
		FailureCount: int64(errorCount), Blobs: blobs}
	JSON(w, 200, &Response{Error:false, Message:"success", Data: response})
}

// DownloadBlobHandler handles download request
// public files are served immediately
// private files are only served to authenticated accounts
func (handler *AppHandler) DownloadBlobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["appName"]
	hash := vars["hash"]

	// more info needed
	if appName == "" || hash == "" {
		BadRequestResponse(w)
		return
	}

	// download file from minio
	// along with the cached blob data
	file, blob, err := handler.repo.DownloadBlob(appName, hash)
	if err != nil {
		JSON(w, 500, &Response{Error:true, Message: "failed to download blob"})
		return
	}

	// not a private file, serve it!
	if !blob.IsPrivate {
		WriteHeaderInfo(w, blob)
		_, err := io.Copy(w, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// its a private file
	// verify caller's private key
	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}

	account, err := handler.store.Get(key[:20])
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	// actually, private files can only
	// be accessed by authenticated accounts.
	// Specifically the account that created them (-_-)
	if account.Cred.PrivateAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	WriteHeaderInfo(w, blob)
	_, err = io.Copy(w, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetAppBlobs get app uploaded files
func (handler *AppHandler) GetAppBlobs(w http.ResponseWriter, r *http.Request) {
	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}

	account, err := handler.store.Get(key[:20])
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	// this resources can only be accessed through
	// privateKey
	cred := account.Cred
	if cred.PrivateAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		BadRequestResponse(w)
		return
	}
	page, err := strconv.Atoi(vars["page"])
	if err != nil {
		BadRequestResponse(w)
		return
	}

	// another level of authorization
	// makes sure the caller is the creator of this app
	app := handler.repo.GetAppByAttr("id", uint(appId))
	if app == nil {
		JSON(w, 404, &Response{Error: true, Message: "app not found"})
		return
	}

	if app.AccountId != account.ID {
		UnAuthorizedResponse(w)
		return
	}

	data := handler.repo.GetAppBlobs(uint(appId), int64(page))
	JSON(w, 200, &Response{Error: false, Message: "success", Data: data})
}

func WriteHeaderInfo(w http.ResponseWriter, blob *models.Blob) {
	headers := map[string]string{"Content-Type": blob.ContentType, "Content-Disposition" :
	fmt.Sprintf("attachment; filename=%s", blob.Filename)}
	for key, val := range headers {
		w.Header().Add(key, val)
	}
}
