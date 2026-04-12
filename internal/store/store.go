package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
	"personstorage/internal/domain"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("Person not found")
var ErrDuplicateExternalID = errors.New("Person with this external_id already exists")
var ErrDuplicateEmail = errors.New("Person with this email already exists")

type Store struct {
	db *gorm.DB
}

func (store *Store) Save(ctx context.Context, person domain.Person) error {
	result := store.db.WithContext(ctx).Create(&person)
	if result.Error != nil {
		if isDuplicateConstraintError(result.Error) && strings.Contains(result.Error.Error(), "people.external_id") {
			fmt.Println(result.Error.Error())
			return ErrDuplicateExternalID
		}
		if isDuplicateConstraintError(result.Error) && strings.Contains(result.Error.Error(), "people.email") {
			fmt.Println(result.Error.Error())
			return ErrDuplicateEmail
		}
		return result.Error
	}

	return nil
}

func (store *Store) Get(ctx context.Context, externalID string) (domain.Person, error) {
	var person domain.Person
	queryCondition := domain.Person{ExternalID: externalID}
	err := store.db.WithContext(ctx).Where(&queryCondition).First(&person).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Person{}, ErrUserNotFound
		}
		return domain.Person{}, err
	}

	return person, nil
}

func isDuplicateConstraintError(err error) bool {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code == sqlite3.ErrConstraint &&
		(sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey ||
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique)
}
