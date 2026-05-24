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
	Store Store
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

		home, err := config.Store.CreateGroup(r.Context(), profile.Player, name)
		if err != nil {
			http.Error(w, "create Group", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, home)
	})
	mux.HandleFunc("GET /v1/groups", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		groups, err := config.Store.ListGroups(r.Context(), profile.Player)
		if err != nil {
			http.Error(w, "list Groups", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, groups)
	})
	mux.HandleFunc("GET /v1/groups/{groupID}/home", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		home, ok, err := config.Store.GroupHome(r.Context(), profile.Player, r.PathValue("groupID"))
		if err != nil {
			http.Error(w, "get Group home", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, home)
	})
	mux.HandleFunc("POST /v1/groups/{groupID}/invites", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		invite, ok, err := config.Store.CreateInvite(r.Context(), profile.Player, r.PathValue("groupID"))
		if err != nil {
			http.Error(w, "create Invite", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, invite)
	})
	mux.HandleFunc("POST /v1/invites/{token}/accept", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		home, status, err := config.Store.AcceptInvite(r.Context(), profile.Player, r.PathValue("token"))
		if err != nil {
			http.Error(w, "accept Invite", http.StatusInternalServerError)
			return
		}
		switch status {
		case InviteInvalid:
			http.Error(w, "Invite cannot be accepted", http.StatusNotFound)
			return
		case InviteUsed:
			http.Error(w, "Invite already used", http.StatusConflict)
			return
		case InviteExpired:
			http.Error(w, "Invite expired", http.StatusGone)
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

	profile, err := config.Store.BootstrapIdentity(r.Context(), identity)
	if err != nil {
		http.Error(w, "bootstrap identity", http.StatusInternalServerError)
		return MeResponse{}, false
	}

	return profile, true
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
