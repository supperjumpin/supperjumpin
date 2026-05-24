package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type AuthIdentity struct {
	Provider string
	Subject  string
	Email    string
}

type AuthVerifier interface {
	Verify(token string) (AuthIdentity, bool)
}

type StaticAuthVerifier map[string]AuthIdentity

func (v StaticAuthVerifier) Verify(token string) (AuthIdentity, bool) {
	identity, ok := v[token]
	return identity, ok
}

type ServerConfig struct {
	Auth  AuthVerifier
	Store *MemoryStore
}

func NewServer(config ServerConfig) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		writeJSON(w, http.StatusOK, profile)
	})
	mux.HandleFunc("POST /v1/groups", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(request.Name)
		if name == "" {
			http.Error(w, "Group name is required", http.StatusBadRequest)
			return
		}

		writeJSON(w, http.StatusCreated, config.Store.CreateGroup(profile.Player, name))
	})
	mux.HandleFunc("GET /v1/groups", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		writeJSON(w, http.StatusOK, config.Store.ListGroups(profile.Player))
	})
	mux.HandleFunc("GET /v1/groups/{groupID}/home", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		home, ok := config.Store.GroupHome(profile.Player, r.PathValue("groupID"))
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, home)
	})
	return mux
}

func signedInProfile(w http.ResponseWriter, r *http.Request, config ServerConfig) (MeResponse, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return MeResponse{}, false
	}

	identity, ok := config.Auth.Verify(token)
	if !ok {
		http.Error(w, "invalid bearer token", http.StatusUnauthorized)
		return MeResponse{}, false
	}

	return config.Store.BootstrapIdentity(identity), true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
	}
}

func bearerToken(header string) (string, bool) {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
