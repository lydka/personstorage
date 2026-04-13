package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"personstorage/internal/domain"
	"personstorage/internal/httpapi"
	"personstorage/internal/store"
)

type stubPersonStore struct {
	upsertFunc        func(context.Context, domain.Person) error
	getFunc           func(context.Context, string) (domain.Person, error)
	lastUpsertPerson  domain.Person
	lastGetExternalID string
	upsertCalls       int
	getCalls          int
}

func (stub *stubPersonStore) Upsert(ctx context.Context, person domain.Person) error {
	stub.upsertCalls++
	stub.lastUpsertPerson = person
	if stub.upsertFunc != nil {
		return stub.upsertFunc(ctx, person)
	}

	return nil
}

func (stub *stubPersonStore) Get(ctx context.Context, externalID string) (domain.Person, error) {
	stub.getCalls++
	stub.lastGetExternalID = externalID
	if stub.getFunc != nil {
		return stub.getFunc(ctx, externalID)
	}

	return domain.Person{}, nil
}

func TestPostByIDIsNotAllowed(testCase *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/people/123e4567-e89b-12d3-a456-426614174000", nil)
	recorder := httptest.NewRecorder()

	httpapi.NewMux(&stubPersonStore{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		testCase.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}

func TestPostSaveReturnsCreatedMessageAndUpsertsPerson(testCase *testing.T) {
	personStore := &stubPersonStore{}
	body := `{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"some name","email":"email@email.com","date_of_birth":"2020-01-01T12:12:34+00:00"}`
	request := httptest.NewRequest(http.MethodPost, "/people", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		testCase.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"message\":\"Successfully saved\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}

	expectedPerson := domain.Person{
		ExternalID:  "123e4567-e89b-12d3-a456-426614174000",
		Name:        "some name",
		Email:       "email@email.com",
		DateOfBirth: "2020-01-01T12:12:34+00:00",
	}
	if personStore.upsertCalls != 1 {
		testCase.Fatalf("expected upsert to be called once, got %d", personStore.upsertCalls)
	}
	if personStore.lastUpsertPerson != expectedPerson {
		testCase.Fatalf("expected upsert person %+v, got %+v", expectedPerson, personStore.lastUpsertPerson)
	}
}

func TestPostSaveRejectsDuplicateEmail(testCase *testing.T) {
	personStore := &stubPersonStore{
		upsertFunc: func(context.Context, domain.Person) error {
			return store.ErrDuplicateEmail
		},
	}
	body := `{"external_id":"223e4567-e89b-12d3-a456-426614174000","name":"other name","email":"email@email.com","date_of_birth":"2021-01-01T12:12:34+00:00"}`
	request := httptest.NewRequest(http.MethodPost, "/people", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		testCase.Fatalf("expected status %d, got %d", http.StatusConflict, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"error\":\"Person with this email already exists\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
}

func TestPostSaveReturnsInternalServerErrorWhenUpsertFails(testCase *testing.T) {
	personStore := &stubPersonStore{
		upsertFunc: func(context.Context, domain.Person) error {
			return errors.New("boom")
		},
	}
	body := `{"external_id":"223e4567-e89b-12d3-a456-426614174000","name":"other name","email":"email@email.com","date_of_birth":"2021-01-01T12:12:34+00:00"}`
	request := httptest.NewRequest(http.MethodPost, "/people", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		testCase.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"error\":\"internal server error\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
}

func TestGetPersonByIDReturnsStoredJSON(testCase *testing.T) {
	personStore := &stubPersonStore{
		getFunc: func(context.Context, string) (domain.Person, error) {
			return domain.Person{
				ExternalID:  "123e4567-e89b-12d3-a456-426614174000",
				Name:        "some name",
				Email:       "email@email.com",
				DateOfBirth: "2020-01-01T12:12:34+00:00",
			}, nil
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/people/123e4567-e89b-12d3-a456-426614174000", nil)
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		testCase.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"external_id\":\"123e4567-e89b-12d3-a456-426614174000\",\"name\":\"some name\",\"email\":\"email@email.com\",\"date_of_birth\":\"2020-01-01T12:12:34+00:00\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
	if personStore.getCalls != 1 {
		testCase.Fatalf("expected get to be called once, got %d", personStore.getCalls)
	}
	if personStore.lastGetExternalID != "123e4567-e89b-12d3-a456-426614174000" {
		testCase.Fatalf("expected get to be called with external ID %q, got %q", "123e4567-e89b-12d3-a456-426614174000", personStore.lastGetExternalID)
	}
}

func TestGetPersonByIDReturnsNotFoundWhenMissing(testCase *testing.T) {
	personStore := &stubPersonStore{
		getFunc: func(context.Context, string) (domain.Person, error) {
			return domain.Person{}, store.ErrUserNotFound
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/people/missing-id", nil)
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		testCase.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"error\":\"Person not found\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
}

func TestGetPersonByIDReturnsInternalServerErrorWhenGetFails(testCase *testing.T) {
	personStore := &stubPersonStore{
		getFunc: func(context.Context, string) (domain.Person, error) {
			return domain.Person{}, errors.New("boom")
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/people/missing-id", nil)
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		testCase.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"error\":\"internal server error\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
}

func TestPostSaveRejectsInvalidJSON(testCase *testing.T) {
	personStore := &stubPersonStore{
		upsertFunc: func(context.Context, domain.Person) error {
			testCase.Fatal("expected upsert not to be called for invalid JSON")
			return nil
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/people", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	httpapi.NewMux(personStore).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		testCase.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}
	expectedBody := "{\"error\":\"invalid json body\"}\n"
	if responseBody := recorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
	if personStore.upsertCalls != 0 {
		testCase.Fatalf("expected upsert not to be called, got %d calls", personStore.upsertCalls)
	}
}
