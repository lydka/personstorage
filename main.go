package main

import (
	"fmt"
	"net/http"
	"os"

	"personstorage/internal/httpapi"
	"personstorage/internal/store"
)

const (
	defaultListenAddr   = ":8080"
	defaultDatabasePath = "data/app.db"
)

type config struct {
	listenAddr   string
	databasePath string
}

func loadConfig() config {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = defaultDatabasePath
	}

	return config{
		listenAddr:   listenAddr,
		databasePath: databasePath,
	}
}

func newMux(databasePath string) (*http.ServeMux, func() error, error) {
	userStore, err := store.NewSQLiteStore(databasePath)
	if err != nil {
		return nil, nil, err
	}

	return httpapi.NewMux(userStore), func() error {
		return userStore.Close()
	}, nil
}

func main() {
	cfg := loadConfig()

	mux, closeStore, err := newMux(cfg.databasePath)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			panic(err)
		}
	}()

	fmt.Printf("Server listening on http://localhost%s\n", cfg.listenAddr)
	if err := http.ListenAndServe(cfg.listenAddr, mux); err != nil {
		panic(err)
	}
}
