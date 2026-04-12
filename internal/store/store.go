package store

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"personstorage/internal/domain"
)

var ErrUserNotFound = errors.New("Person not found")
var ErrDuplicateEmail = errors.New("Person with this email already exists")

type Store struct {
	db *gorm.DB
}

func (store *Store) Get(requestContext context.Context, externalID string) (domain.Person, error) {
	var person domain.Person
	queryCondition := domain.Person{ExternalID: externalID}
	err := store.db.WithContext(requestContext).Where(&queryCondition).First(&person).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Person{}, ErrUserNotFound
		}
		return domain.Person{}, err
	}

	return person, nil
}

func (store *Store) Save(requestContext context.Context, input domain.Person) error {
	return store.db.WithContext(requestContext).Transaction(func(databaseTransaction *gorm.DB) error {
		person, err := findPerson(databaseTransaction, input.ExternalID)
		if err != nil {
			return err
		}
		if person != nil {
			return updatePerson(databaseTransaction, person, input)
		}

		return createPerson(databaseTransaction, input)
	})
}

func updatePerson(databaseTransaction *gorm.DB, storedPerson *domain.Person, input domain.Person) error {
	storedPerson.Name = input.Name
	storedPerson.Email = input.Email
	storedPerson.DateOfBirth = input.DateOfBirth

	result := databaseTransaction.Save(storedPerson)
	if result.Error != nil {
		if isDuplicateEmailUniqueConstraintError(result.Error) {
			return ErrDuplicateEmail
		}
		return result.Error
	}

	return nil
}

func createPerson(databaseTransaction *gorm.DB, input domain.Person) error {
	result := databaseTransaction.Create(&input)
	if result.Error != nil {
		if isDuplicateEmailUniqueConstraintError(result.Error) {
			return ErrDuplicateEmail
		}
		return result.Error
	}

	return nil
}

func findPerson(databaseTransaction *gorm.DB, externalID string) (*domain.Person, error) {
	var person domain.Person
	queryCondition := domain.Person{ExternalID: externalID}
	err := databaseTransaction.Where(&queryCondition).First(&person).Error
	if err == nil {
		return &person, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return nil, err
}

func isDuplicateEmailUniqueConstraintError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed: people.email")
}
