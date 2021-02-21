package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ConnectSqliteDatabase connects to a given the SQLite database; optionally, applies migration.
// Keep in mind that this is obviously for a non-production purpose.
func ConnectSqliteDatabase(databaseFilePath string, autoMigrate bool) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(databaseFilePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database at '%s' due to: %s", databaseFilePath, err)
	}

	log.Printf("Connected to database at '%s'", databaseFilePath)

	if autoMigrate {
		err := db.AutoMigrate(&Notification{})
		if err != nil {
			return nil, fmt.Errorf("failed to apply migration to database at '%s' due to: %s", databaseFilePath, err)
		}

		log.Printf("Migration applied to database at '%s'", databaseFilePath)
	}

	return db, nil
}
