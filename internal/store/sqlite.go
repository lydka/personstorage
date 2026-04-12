package store

import (
	"os"
	"path/filepath"

	"personstorage/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewSQLiteStore(databasePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&domain.Person{}); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (store *Store) Close() error {
	sqlDatabase, err := store.db.DB()
	if err != nil {
		return err
	}

	return sqlDatabase.Close()
}
