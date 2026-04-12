package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"personstorage/internal/domain"
	"personstorage/internal/store"
)

type saveResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func saveHandler(personStore personStore, responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var payload domain.Person
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeJSON(responseWriter, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	if err := personStore.Save(request.Context(), payload); err != nil {
		if errors.Is(err, store.ErrDuplicateEmail) {
			writeJSON(responseWriter, http.StatusConflict, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(responseWriter, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(responseWriter, http.StatusCreated, saveResponse{Message: "Successfully saved"})
}

func loadHandler(personStore personStore, responseWriter http.ResponseWriter, request *http.Request) {
	response, err := personStore.Get(request.Context(), request.PathValue("id"))
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeJSON(responseWriter, http.StatusNotFound, errorResponse{Error: store.ErrUserNotFound.Error()})
			return
		}

		writeJSON(responseWriter, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(responseWriter, http.StatusOK, response)
}

func writeJSON(responseWriter http.ResponseWriter, statusCode int, payload any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	if err := json.NewEncoder(responseWriter).Encode(payload); err != nil {
		http.Error(responseWriter, "internal server error", http.StatusInternalServerError)
	}
}
