package store

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"personstorage/internal/domain"
)

type sqlDatabaseCloser interface {
	Close() error
}

type sqliteDependencies struct {
	makeDirectoryAll func(string, os.FileMode) error
	openSQLiteDB     func(string) (*gorm.DB, error)
	autoMigrate      func(*gorm.DB) error
	getSQLDatabase   func(*gorm.DB) (sqlDatabaseCloser, error)
}

func defaultSQLiteDependencies() sqliteDependencies {
	return sqliteDependencies{
		makeDirectoryAll: os.MkdirAll,
		openSQLiteDB: func(databasePath string) (*gorm.DB, error) {
			return gorm.Open(sqlite.Open(databasePath), &gorm.Config{
				Logger: gormlogger.Default.LogMode(gormlogger.Silent),
			})
		},
		autoMigrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&domain.Person{})
		},
		getSQLDatabase: func(database *gorm.DB) (sqlDatabaseCloser, error) {
			return database.DB()
		},
	}
}

func NewSQLiteStore(databasePath string) (*Store, error) {
	return newSQLiteStoreWithDependencies(databasePath, defaultSQLiteDependencies())
}

func newSQLiteStoreWithDependencies(databasePath string, dependencies sqliteDependencies) (*Store, error) {
	if err := dependencies.makeDirectoryAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, err
	}

	db, err := dependencies.openSQLiteDB(databasePath)
	if err != nil {
		return nil, err
	}

	if err := dependencies.autoMigrate(db); err != nil {
		return nil, err
	}

	return &Store{
		db:             db,
		getSQLDatabase: dependencies.getSQLDatabase,
	}, nil
}

func (store *Store) Close() error {
	getSQLDatabase := store.getSQLDatabase
	if getSQLDatabase == nil {
		getSQLDatabase = defaultSQLiteDependencies().getSQLDatabase
	}

	sqlDatabase, err := getSQLDatabase(store.db)
	if err != nil {
		return err
	}

	return sqlDatabase.Close()
}
