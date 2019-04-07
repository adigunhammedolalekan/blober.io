package main

import (
	"blober.io/handlers"
	"blober.io/repos"
	"blober.io/services"
	"blober.io/store"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {

	// load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load env variables, %v", err)
	}

	// connect to database or quit app
	db, err := services.CreateDatabaseConnection(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database, %v", err)
	}

	// initialize/open badgerDB backed session store
	sessionStore, err := store.NewSessionStore()
	if err != nil {
		log.Fatalf("failed to open session store %v, quiting...", err)
	}

	// initialize/open badgerDB backed blob struct cache
	blobStore, err := store.NewBlobStore()
	if err != nil {
		log.Fatalf("failed to open blob store, %v", err)
	}

	defer func() {

		log.Println("Closing app...")
		err := db.Close()
		if err != nil {
			log.Printf("failed to close DB, %v", err)
		}

		err = sessionStore.Close()
		if err != nil {
			log.Printf("failed to close session store, %v", err)
		}

		err = blobStore.Close()
		if err != nil {
			log.Printf("failed to close blob store, %v", err)
		}
	}()

	// open and connect to minio object storage server
	storage, err := services.NewStorageService(&services.Option{
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Host: os.Getenv("MINIO_HOST"),
		Store:blobStore,
	})

	if err != nil {
		log.Fatalf("failed to start minio server %v", err)
	}

	// create app dependencies
	accountRepo := repos.NewAccountRepository(sessionStore, db)
	appRepo := repos.NewAppRepository(db, accountRepo, storage)
	accountHandler := handlers.NewAccountHandler(accountRepo)
	appHandler := handlers.NewAppHandler(sessionStore, blobStore, appRepo)

	router := mux.NewRouter()
	router.NotFoundHandler = &handlers.NotFoundHandler{}
	router.MethodNotAllowedHandler = &handlers.MethodNotAllowedHandler{}

	router.HandleFunc("/account/new", accountHandler.CreateNewAccountHandler).Methods("POST")
	router.HandleFunc("/account/authenticate", accountHandler.AuthenticateAccountHandler).Methods("POST")
	router.HandleFunc("/app/new", appHandler.CreateNewAppHandler).Methods("POST")
	router.HandleFunc("/me/apps", appHandler.GetAccountAppsHandler).Methods("GET")
	router.HandleFunc("/{appName}/upload", appHandler.UploadBlobHandler).Methods("POST")
	router.HandleFunc("/{appName}/uploads", appHandler.UploadMultipleBlobsHandler).Methods("POST")
	router.HandleFunc("/{appName}/{hash}/download", appHandler.DownloadBlobHandler).Methods("GET")
	router.HandleFunc("/apps/{appId}/blobs/{page}", appHandler.GetAppBlobs).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "9008"
	}

	address := "0.0.0.0:" + port
	fmt.Println("Server started at ", address)

	handler := cors.Default().Handler(router)
	if err := http.ListenAndServe(address, handler); err != nil {
		log.Fatal(err)
	}
}
