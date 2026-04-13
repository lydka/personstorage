package httpapi

import (
	"context"
	"personstorage/internal/domain"
	"net/http"
)

type personStore interface {
	Upsert(ctx context.Context, person domain.Person) error
	Get(ctx context.Context, externalID string) (domain.Person, error)
}

func NewMux(personStore personStore) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /people", func(responseWriter http.ResponseWriter, request *http.Request) {
		saveHandler(personStore, responseWriter, request)
	})
	mux.HandleFunc("GET /people/{id}", func(responseWriter http.ResponseWriter, request *http.Request) {
		loadHandler(personStore, responseWriter, request)
	})
	return mux
}
