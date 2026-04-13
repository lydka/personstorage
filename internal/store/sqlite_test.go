package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubSQLDatabase struct {
	closeErr error
}

func (database stubSQLDatabase) Close() error {
	return database.closeErr
}

func TestNewSQLiteStoreReturnsErrorWhenDirectoryCreationFails(testCase *testing.T) {
	expectedErr := errors.New("mkdir failed")
	personStore, err := newSQLiteStoreWithDependencies(filepath.Join("ignored", "test.db"), sqliteDependencies{
		makeDirectoryAll: func(string, os.FileMode) error {
			return expectedErr
		},
		openSQLiteDB: func(string) (*gorm.DB, error) {
			testCase.Fatal("expected openSQLiteDB not to be called")
			return nil, nil
		},
		autoMigrate: func(*gorm.DB) error {
			testCase.Fatal("expected autoMigrate not to be called")
			return nil
		},
		getSQLDatabase: func(*gorm.DB) (sqlDatabaseCloser, error) {
			testCase.Fatal("expected getSQLDatabase not to be called")
			return nil, nil
		},
	})
	if !errors.Is(err, expectedErr) {
		testCase.Fatalf("expected mkdir error %v, got %v", expectedErr, err)
	}
	if personStore != nil {
		testCase.Fatalf("expected nil store, got %+v", personStore)
	}
}

func TestNewSQLiteStoreReturnsErrorWhenOpenFails(testCase *testing.T) {
	expectedErr := errors.New("open failed")
	personStore, err := newSQLiteStoreWithDependencies(filepath.Join("ignored", "test.db"), sqliteDependencies{
		makeDirectoryAll: func(string, os.FileMode) error {
			return nil
		},
		openSQLiteDB: func(string) (*gorm.DB, error) {
			return nil, expectedErr
		},
		autoMigrate: func(*gorm.DB) error {
			testCase.Fatal("expected autoMigrate not to be called")
			return nil
		},
		getSQLDatabase: func(*gorm.DB) (sqlDatabaseCloser, error) {
			testCase.Fatal("expected getSQLDatabase not to be called")
			return nil, nil
		},
	})
	if !errors.Is(err, expectedErr) {
		testCase.Fatalf("expected open error %v, got %v", expectedErr, err)
	}
	if personStore != nil {
		testCase.Fatalf("expected nil store, got %+v", personStore)
	}
}

func TestNewSQLiteStoreReturnsErrorWhenMigrationFails(testCase *testing.T) {
	expectedErr := errors.New("migration failed")
	database := newSQLiteDatabase(testCase)
	personStore, err := newSQLiteStoreWithDependencies(filepath.Join("ignored", "test.db"), sqliteDependencies{
		makeDirectoryAll: func(string, os.FileMode) error {
			return nil
		},
		openSQLiteDB: func(string) (*gorm.DB, error) {
			return database, nil
		},
		autoMigrate: func(*gorm.DB) error {
			return expectedErr
		},
		getSQLDatabase: func(*gorm.DB) (sqlDatabaseCloser, error) {
			return stubSQLDatabase{}, nil
		},
	})
	if !errors.Is(err, expectedErr) {
		testCase.Fatalf("expected migration error %v, got %v", expectedErr, err)
	}
	if personStore != nil {
		testCase.Fatalf("expected nil store, got %+v", personStore)
	}
}

func TestNewSQLiteStoreCreatesDatabasePathAndCloses(testCase *testing.T) {
	databasePath := filepath.Join(testCase.TempDir(), "nested", "test.db")

	personStore, err := NewSQLiteStore(databasePath)
	if err != nil {
		testCase.Fatalf("expected store creation to succeed, got %v", err)
	}

	if _, err := os.Stat(databasePath); err != nil {
		testCase.Fatalf("expected database file to exist at %s, got %v", databasePath, err)
	}

	if err := personStore.Close(); err != nil {
		testCase.Fatalf("expected close to succeed, got %v", err)
	}
}

func TestCloseReturnsErrorWhenSQLDatabaseLookupFails(testCase *testing.T) {
	expectedErr := errors.New("db lookup failed")
	personStore := &Store{
		getSQLDatabase: func(*gorm.DB) (sqlDatabaseCloser, error) {
			return nil, expectedErr
		},
	}

	err := personStore.Close()
	if !errors.Is(err, expectedErr) {
		testCase.Fatalf("expected db lookup error %v, got %v", expectedErr, err)
	}
}

func TestCloseReturnsErrorWhenSQLDatabaseCloseFails(testCase *testing.T) {
	expectedErr := errors.New("close failed")
	personStore := &Store{
		getSQLDatabase: func(*gorm.DB) (sqlDatabaseCloser, error) {
			return stubSQLDatabase{closeErr: expectedErr}, nil
		},
	}

	err := personStore.Close()
	if !errors.Is(err, expectedErr) {
		testCase.Fatalf("expected close error %v, got %v", expectedErr, err)
	}
}

func newSQLiteDatabase(testCase *testing.T) *gorm.DB {
	testCase.Helper()

	database, err := gorm.Open(sqlite.Open(filepath.Join(testCase.TempDir(), "unit.db")), &gorm.Config{})
	if err != nil {
		testCase.Fatalf("expected sqlite database to open, got %v", err)
	}

	return database
}
