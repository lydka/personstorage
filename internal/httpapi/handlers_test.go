package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"personstorage/internal/httpapi"
	"personstorage/internal/store"
)

func newTestMux(testCase *testing.T) *http.ServeMux {
	testCase.Helper()

	userStore, err := store.NewSQLiteStore(filepath.Join(testCase.TempDir(), "test.db"))
	if err != nil {
		testCase.Fatalf("failed to create test store: %v", err)
	}
	testCase.Cleanup(func() {
		if err := userStore.Close(); err != nil {
			testCase.Fatalf("failed to close test store: %v", err)
		}
	})

	return httpapi.NewMux(userStore)
}

func TestPostByIDIsNotAllowed(testCase *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/123e4567-e89b-12d3-a456-426614174000", nil)
	recorder := httptest.NewRecorder()

	newTestMux(testCase).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		testCase.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}

func TestPostSaveReturnsCreatedMessage(testCase *testing.T) {
	body := `{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"some name","email":"email@email.com","date_of_birth":"2020-01-01T12:12:34+00:00"}`
	request := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	newTestMux(testCase).ServeHTTP(recorder, request)

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
}

func TestPostSaveRejectsDuplicateExternalID(testCase *testing.T) {
	mux := newTestMux(testCase)
	firstBody := `{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"some name","email":"email@email.com","date_of_birth":"2020-01-01T12:12:34+00:00"}`
	secondBody := `{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"other name","email":"other@email.com","date_of_birth":"2021-01-01T12:12:34+00:00"}`

	firstRequest := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(firstBody))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRecorder := httptest.NewRecorder()
	mux.ServeHTTP(firstRecorder, firstRequest)

	secondRequest := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(secondBody))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRecorder := httptest.NewRecorder()
	mux.ServeHTTP(secondRecorder, secondRequest)

	if secondRecorder.Code != http.StatusConflict {
		testCase.Fatalf("expected status %d, got %d", http.StatusConflict, secondRecorder.Code)
	}

	if contentType := secondRecorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}

	expectedBody := "{\"error\":\"Person with this external_id already exists\"}\n"
	if responseBody := secondRecorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
}

func TestPostSaveRejectsDuplicateEmail(testCase *testing.T) {
	mux := newTestMux(testCase)
	firstBody := `{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"some name","email":"email@email.com","date_of_birth":"2020-01-01T12:12:34+00:00"}`
	secondBody := `{"external_id":"223e4567-e89b-12d3-a456-426614174000","name":"other name","email":"email@email.com","date_of_birth":"2021-01-01T12:12:34+00:00"}`

	firstRequest := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(firstBody))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRecorder := httptest.NewRecorder()
	mux.ServeHTTP(firstRecorder, firstRequest)

	secondRequest := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(secondBody))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRecorder := httptest.NewRecorder()
	mux.ServeHTTP(secondRecorder, secondRequest)

	if secondRecorder.Code != http.StatusConflict {
		testCase.Fatalf("expected status %d, got %d", http.StatusConflict, secondRecorder.Code)
	}

	if contentType := secondRecorder.Header().Get("Content-Type"); contentType != "application/json" {
		testCase.Fatalf("expected content type %q, got %q", "application/json", contentType)
	}

	expectedBody := "{\"error\":\"Person with this email already exists\"}\n"
	if responseBody := secondRecorder.Body.String(); responseBody != expectedBody {
		testCase.Fatalf("expected body %q, got %q", expectedBody, responseBody)
	}
}

func TestGetByIDReturnsStoredJSON(testCase *testing.T) {
	mux := newTestMux(testCase)
	body := `{"external_id":"123e4567-e89b-12d3-a456-426614174000","name":"some name","email":"email@email.com","date_of_birth":"2020-01-01T12:12:34+00:00"}`

	saveRequest := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(body))
	saveRequest.Header.Set("Content-Type", "application/json")
	saveRecorder := httptest.NewRecorder()
	mux.ServeHTTP(saveRecorder, saveRequest)

	request := httptest.NewRequest(http.MethodGet, "/123e4567-e89b-12d3-a456-426614174000", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

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
}

func TestGetByIDReturnsNotFoundWhenMissing(testCase *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/missing-id", nil)
	recorder := httptest.NewRecorder()

	newTestMux(testCase).ServeHTTP(recorder, request)

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

func TestPostSaveRejectsInvalidJSON(testCase *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	newTestMux(testCase).ServeHTTP(recorder, request)

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
}
