package services

import (
	"blober.io/models"
	"github.com/jinzhu/gorm"
)
import _ "github.com/jinzhu/gorm/dialects/postgres"

// CreateDatabaseConnection creates connection
// to a PostgresDB and performs migration
func CreateDatabaseConnection(connectUri string) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", connectUri)
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&models.Account{}, &models.App{}, &models.Blob{})
	return db, err
}
