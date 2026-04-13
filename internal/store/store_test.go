package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"personstorage/internal/domain"
)

func TestIsDuplicateEmailUniqueConstraintError(testCase *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "email unique constraint",
			err:  errors.New("UNIQUE constraint failed: people.email"),
			want: true,
		},
		{
			name: "external id unique constraint",
			err:  errors.New("UNIQUE constraint failed: people.external_id"),
			want: false,
		},
		{
			name: "non sqlite error",
			err:  errors.New("something else failed"),
			want: false,
		},
	}

	for _, test := range testCases {
		testCase.Run(test.name, func(testCase *testing.T) {
			if got := isDuplicateEmailUniqueConstraintError(test.err); got != test.want {
				testCase.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}

func TestUpsertCreatesPerson(testCase *testing.T) {
	personStore := newTestStore(testCase)
	person := samplePerson()

	if err := personStore.Upsert(context.Background(), person); err != nil {
		testCase.Fatalf("expected upsert to succeed, got %v", err)
	}

	storedPerson, err := personStore.Get(context.Background(), person.ExternalID)
	if err != nil {
		testCase.Fatalf("expected get to succeed, got %v", err)
	}

	assertPersonEqual(testCase, storedPerson, person)
}

func TestUpsertUpdatesExistingPerson(testCase *testing.T) {
	personStore := newTestStore(testCase)
	originalPerson := samplePerson()
	updatedPerson := domain.Person{
		ExternalID:  originalPerson.ExternalID,
		Name:        "Updated Name",
		Email:       "updated@example.com",
		DateOfBirth: "2001-02-03",
	}

	if err := personStore.Upsert(context.Background(), originalPerson); err != nil {
		testCase.Fatalf("expected initial upsert to succeed, got %v", err)
	}

	if err := personStore.Upsert(context.Background(), updatedPerson); err != nil {
		testCase.Fatalf("expected update upsert to succeed, got %v", err)
	}

	storedPerson, err := personStore.Get(context.Background(), updatedPerson.ExternalID)
	if err != nil {
		testCase.Fatalf("expected get to succeed, got %v", err)
	}

	assertPersonEqual(testCase, storedPerson, updatedPerson)
}

func TestUpsertRejectsDuplicateEmailForDifferentPerson(testCase *testing.T) {
	personStore := newTestStore(testCase)
	firstPerson := samplePerson()
	secondPerson := domain.Person{
		ExternalID:  "person-2",
		Name:        "Second Person",
		Email:       firstPerson.Email,
		DateOfBirth: "1999-09-09",
	}

	if err := personStore.Upsert(context.Background(), firstPerson); err != nil {
		testCase.Fatalf("expected initial upsert to succeed, got %v", err)
	}

	err := personStore.Upsert(context.Background(), secondPerson)
	if !errors.Is(err, ErrDuplicateEmail) {
		testCase.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestGetReturnsNotFoundForMissingPerson(testCase *testing.T) {
	personStore := newTestStore(testCase)

	_, err := personStore.Get(context.Background(), "missing-external-id")
	if !errors.Is(err, ErrUserNotFound) {
		testCase.Fatalf("expected not found error, got %v", err)
	}
}

func TestFindPersonReturnsStoredPerson(testCase *testing.T) {
	personStore := newTestStore(testCase)
	person := samplePerson()

	if err := personStore.Upsert(context.Background(), person); err != nil {
		testCase.Fatalf("expected upsert to succeed, got %v", err)
	}

	storedPerson, err := findPerson(personStore.db, person.ExternalID)
	if err != nil {
		testCase.Fatalf("expected findPerson to succeed, got %v", err)
	}
	if storedPerson == nil {
		testCase.Fatal("expected findPerson to return a person, got nil")
	}

	assertPersonEqual(testCase, *storedPerson, person)
}

func TestFindPersonReturnsNilWhenMissing(testCase *testing.T) {
	personStore := newTestStore(testCase)

	person, err := findPerson(personStore.db, "missing-external-id")
	if err != nil {
		testCase.Fatalf("expected findPerson to succeed, got %v", err)
	}
	if person != nil {
		testCase.Fatalf("expected nil person, got %+v", *person)
	}
}

func newTestStore(testCase *testing.T) *Store {
	testCase.Helper()

	personStore, err := NewSQLiteStore(filepath.Join(testCase.TempDir(), "test.db"))
	if err != nil {
		testCase.Fatalf("failed to create test store: %v", err)
	}

	testCase.Cleanup(func() {
		if err := personStore.Close(); err != nil {
			testCase.Fatalf("failed to close test store: %v", err)
		}
	})

	return personStore
}

func samplePerson() domain.Person {
	return domain.Person{
		ExternalID:  "person-1",
		Name:        "Sample Person",
		Email:       "sample@example.com",
		DateOfBirth: "2000-01-02",
	}
}

func assertPersonEqual(testCase *testing.T, got domain.Person, want domain.Person) {
	testCase.Helper()

	if got != want {
		testCase.Fatalf("expected person %+v, got %+v", want, got)
	}
}
