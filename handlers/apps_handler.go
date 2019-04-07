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

var maxMemory int64 = 1024 << 20

type AppHandler struct {
	store *store.SessionStore
	blobStore *store.BlobStore
	repo *repos.AppRepository
}

func NewAppHandler(store *store.SessionStore, blobStore *store.BlobStore, repo *repos.AppRepository) *AppHandler {
	return &AppHandler{repo:repo, blobStore:blobStore, store:store}
}

func (handler *AppHandler) CreateNewAppHandler(w http.ResponseWriter, r *http.Request) {

	key := ParseAuthorizationKey(r)
	if key == "" || len(key) < 20 {
		UnAuthorizedResponse(w)
		return
	}

	sessionKey := key[:20]
	account, err := handler.store.Get(sessionKey)
	if err != nil {
		UnAuthorizedResponse(w)
		return
	}

	cred := account.Cred
	if cred.PrivateAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

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

	vars := mux.Vars(r)
	appName := vars["appName"]
	if appName == "" {
		BadRequestResponse(w)
		return
	}

	private := r.FormValue("private")
	isPrivate := private == "true"
	blob, err := handler.repo.UploadBlob(account.ID, appName, isPrivate, header)
	if err != nil {
		JSON(w, 200, &Response{Error: true, Message: err.Error()})
		return
	}

	JSON(w, 200, &Response{Error: false, Message: "success", Data: blob})
}

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

	cred := account.Cred
	if cred.PrivateAccessKey != key && cred.PublicAccessKey != key {
		UnAuthorizedResponse(w)
		return
	}

	vars := mux.Vars(r)
	appName := vars["appName"]
	if appName == "" {
		BadRequestResponse(w)
		return
	}

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		BadRequestResponse(w)
		return
	}

	errorCount := 0
	successCount := 0

	blobs := make([]*models.Blob, 0)
	files := r.MultipartForm.File["files[]"]
	if len(files) == 0 {
		JSON(w, 400, &Response{Error: true, Message: "no file found"})
		return
	}

	isPrivate := r.FormValue("private") == "true"
	for _, value := range files {

		blob, err := handler.repo.UploadBlob(account.ID, appName, isPrivate, value)
		if err != nil {
			log.Printf("failed to process upload, %v", err)
			errorCount += 1
			continue
		}

		blobs = append(blobs, blob)
		successCount += 1
	}

	response := &models.UploadMultipleResponse{SuccessCount:int64(successCount),
		FailureCount: int64(errorCount), Blobs: blobs}
	JSON(w, 200, &Response{Error:false, Message:"success", Data: response})
}

func (handler *AppHandler) DownloadBlobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["appName"]
	hash := vars["hash"]

	if appName == "" || hash == "" {
		BadRequestResponse(w)
		return
	}

	file, blob, err := handler.repo.DownloadBlob(appName, hash)
	if err != nil {
		JSON(w, 500, &Response{Error:true, Message: "failed to download blob"})
		return
	}

	if !blob.IsPrivate {
		WriteHeaderInfo(w, blob)
		_, err := io.Copy(w, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

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
	JSON(w, 200, &Response{Error: false, Message:"success", Data:data})
}

func WriteHeaderInfo(w http.ResponseWriter, blob *models.Blob) {
	headers := map[string]string{"Content-Type": blob.ContentType, "Content-Disposition" :
	fmt.Sprintf("attachment; filename=%s", blob.Filename)}
	for key, val := range headers {
		w.Header().Add(key, val)
	}
}