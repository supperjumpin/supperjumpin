package main

import (
	"log"
	"net/http"
	"os"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	auth := httpapi.StaticAuthVerifier{}
	if token := os.Getenv("SUPPERJUMPIN_DEV_AUTH_TOKEN"); token != "" {
		auth[token] = httpapi.AuthIdentity{
			Provider: "supabase",
			Subject:  envOrDefault("SUPPERJUMPIN_DEV_AUTH_SUBJECT", "dev-supabase-subject"),
			Email:    envOrDefault("SUPPERJUMPIN_DEV_AUTH_EMAIL", "player@example.com"),
		}
	}

	server := httpapi.NewServer(httpapi.ServerConfig{Auth: auth, Store: httpapi.NewMemoryStore()})
	log.Printf("Supperjumpin API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
