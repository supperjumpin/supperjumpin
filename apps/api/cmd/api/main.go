package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	auth := httpapi.AuthVerifierChain{}
	if token := os.Getenv("SUPPERJUMPIN_DEV_AUTH_TOKEN"); token != "" {
		auth = append(auth, httpapi.StaticAuthVerifier{token: {
			Provider: "local-dev",
			Subject:  envOrDefault("SUPPERJUMPIN_DEV_AUTH_SUBJECT", "dev-player"),
			Email:    envOrDefault("SUPPERJUMPIN_DEV_AUTH_EMAIL", "player@example.com"),
		}})
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required for durable Supperjumpin API state")
	}
	store, err := httpapi.NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("connect to Postgres: %v", err)
	}
	defer store.Close()

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth:         auth,
		Store:        store,
		Now:          time.Now,
		JumpPlanning: store,
		Judgment:     store,
		PublicRead:   store,
		Open:         store,
		CaptionEdit:  store,
		JumpRetract:  store,
	})
	log.Printf("Supperjumpin API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
