package httpapi

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

type AuthIdentity struct {
	Provider string
	Subject  string
	Email    string
}

type AuthVerifier interface {
	Verify(token string) (AuthIdentity, bool)
}

type ServerConfig struct {
	Auth   AuthVerifier
	Store  Store
	Now    func() time.Time
	Logger *slog.Logger
}

func NewServer(config ServerConfig) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/me", "bootstrap_identity")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		writeJSON(w, http.StatusOK, profile)
	})
	mux.HandleFunc("PATCH /v1/me/display-name", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "PATCH /v1/me/display-name", "update_display_name")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			DisplayName string `json:"displayName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		displayName := strings.TrimSpace(request.DisplayName)
		if displayName == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_display_name", nil)
			http.Error(w, "displayName is required", http.StatusBadRequest)
			return
		}

		player, err := config.Store.UpdateDisplayName(r.Context(), profile.Player.ID, displayName)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "update_display_name_failed", err)
			http.Error(w, "update display name", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, UpdateDisplayNameResponse{Player: player})
	})
	mux.HandleFunc("GET /v1/prompt-catalog", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/prompt-catalog", "list_prompt_catalog")
		AddRequestLogField(r.Context(), "actor_type", "public")

		result, err := game.ListCatalog(r.Context(), config.Store)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "list_prompt_catalog_failed", err)
			http.Error(w, "list prompt catalog", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusInternalServerError, "list_prompt_catalog_failed", result.Err)
			http.Error(w, "list prompt catalog", http.StatusInternalServerError)
			return
		}

		packs := make([]PromptPackDTO, 0, len(result.Catalog.Packs))
		for _, pw := range result.Catalog.Packs {
			prompts := make([]PromptDTO, 0, len(pw.Prompts))
			for _, p := range pw.Prompts {
				prompts = append(prompts, PromptDTO{
					ID:       p.ID,
					Copy:     p.Copy,
					Theme:    p.Theme,
					CostTier: p.CostTier,
				})
			}
			packs = append(packs, PromptPackDTO{
				ID:          pw.Pack.ID,
				DisplayName: pw.Pack.DisplayName,
				Description: pw.Pack.Description,
				Prompts:     prompts,
			})
		}

		writeJSON(w, http.StatusOK, PromptCatalogResponse{Packs: packs})
	})
	mux.HandleFunc("GET /v1/reveal-timeframes", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/reveal-timeframes", "list_reveal_timeframes")
		AddRequestLogField(r.Context(), "actor_type", "public")

		result, err := game.ListRevealTimeframes(r.Context(), config.Store)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "list_reveal_timeframes_failed", err)
			http.Error(w, "list reveal timeframes", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusInternalServerError, "list_reveal_timeframes_failed", result.Err)
			http.Error(w, "list reveal timeframes", http.StatusInternalServerError)
			return
		}

		tfs := make([]RevealTimeframeDTO, 0, len(result.Timeframes))
		for _, tf := range result.Timeframes {
			tfs = append(tfs, RevealTimeframeDTO{
				ID:            tf.ID,
				Label:         tf.Label,
				DurationHours: tf.DurationHours,
			})
		}

		writeJSON(w, http.StatusOK, RevealTimeframesResponse{Timeframes: tfs})
	})
	mux.HandleFunc("POST /v1/rounds", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds", "start_round")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request StartRoundRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if request.CommunityID == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_community_id", nil)
			http.Error(w, "communityId is required", http.StatusBadRequest)
			return
		}
		if request.RevealTimeframeID == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_reveal_timeframe_id", nil)
			http.Error(w, "revealTimeframeId is required", http.StatusBadRequest)
			return
		}

		now := config.Now()
		result, err := game.StartRound(r.Context(), config.Store, game.StartRoundInput{
			CommunityID:       request.CommunityID,
			PlayerID:          profile.Player.ID,
			PromptID:          request.PromptID,
			RevealTimeframeID: request.RevealTimeframeID,
		}, now, rand.Intn)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "start_round_failed", err)
			http.Error(w, "start round", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "start_round_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		round := RoundDTO{
			ID:          result.Round.ID,
			CommunityID: result.Round.CommunityID,
			PromptID:    result.Round.PromptID,
			Status:      result.Round.Status,
			RevealBy:    result.Round.RevealBy.Format(time.RFC3339),
			CreatedBy:   result.Round.CreatedBy,
			CreatedAt:   result.Round.CreatedAt.Format(time.RFC3339),
		}

		writeJSON(w, http.StatusCreated, StartRoundResponse{Round: round})
	})

	return requestLoggingMiddleware(mux, config.Logger, config.Now)
}

func signedInProfile(w http.ResponseWriter, r *http.Request, config ServerConfig) (MeResponse, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		recordHTTPError(r, http.StatusUnauthorized, "missing_bearer_token", nil)
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return MeResponse{}, false
	}

	identity, ok := config.Auth.Verify(token)
	if !ok {
		recordHTTPError(r, http.StatusUnauthorized, "invalid_bearer_token", nil)
		http.Error(w, "invalid bearer token", http.StatusUnauthorized)
		return MeResponse{}, false
	}

	profile, err := config.Store.BootstrapIdentity(r.Context(), identity)
	if err != nil {
		recordHTTPError(r, http.StatusInternalServerError, "bootstrap_identity_failed", err)
		http.Error(w, "bootstrap identity", http.StatusInternalServerError)
		return MeResponse{}, false
	}
	AddRequestLogFields(r.Context(), slog.String("actor_type", "player"), slog.String("player_id", profile.Player.ID))

	return profile, true
}

func setRequestOperation(r *http.Request, route string, operation string) {
	AddRequestLogFields(r.Context(), slog.String("route", route), slog.String("operation", operation))
}

func recordHTTPError(r *http.Request, status int, code string, err error) {
	AddRequestLogFields(r.Context(), slog.String("outcome", outcomeForStatus(status)), slog.String("error_code", code))
	if status >= http.StatusInternalServerError {
		RaiseRequestLogLevel(r.Context(), slog.LevelError)
		AddRequestLogField(r.Context(), "stack", string(debug.Stack()))
		if err != nil {
			AddRequestLogField(r.Context(), "internal_error", safeInternalError(err))
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}

func bearerToken(header string) (string, bool) {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
