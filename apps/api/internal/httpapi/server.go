package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	Auth         AuthVerifier
	Store        Store
	Now          func() time.Time
	JumpPlanning JumpPlanningFlow
	Judgment     JudgmentFlow
	PublicRead   PublicReadFlow
	Open         OpenFlow
	CaptionEdit  CaptionEditFlow
	JumpRetract  JumpRetractFlow
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
	mux.HandleFunc("PATCH /v1/me/display-name", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			DisplayName string `json:"displayName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		displayName := strings.TrimSpace(request.DisplayName)
		if displayName == "" {
			http.Error(w, "displayName is required", http.StatusBadRequest)
			return
		}

		player, err := config.Store.UpdateDisplayName(r.Context(), profile.Player.ID, displayName)
		if err != nil {
			http.Error(w, "update display name", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, UpdateDisplayNameResponse{Player: player})
	})
	mux.HandleFunc("POST /v1/jumps", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Source         string `json:"source"`
			Destination    string `json:"destination"`
			Food           string `json:"food"`
			Caption        string `json:"caption"`
			MediaObjectKey string `json:"mediaObjectKey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		source := strings.TrimSpace(request.Source)
		destination := strings.TrimSpace(request.Destination)
		food := strings.TrimSpace(request.Food)
		caption := strings.TrimSpace(request.Caption)
		mediaObjectKey := strings.TrimSpace(request.MediaObjectKey)
		if source == "" || destination == "" || food == "" || caption == "" || mediaObjectKey == "" {
			http.Error(w, "Source, Destination, Food, Caption, and mediaObjectKey are required", http.StatusBadRequest)
			return
		}

		jump, err := createPerformedJump(r.Context(), config.JumpPlanning, profile.Player, source, destination, food, caption, mediaObjectKey, config.Now())
		if err != nil {
			http.Error(w, "create Performed Jump", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, jump)
	})
	mux.HandleFunc("PATCH /v1/jumps/{jumpID}", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Caption string `json:"caption"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		caption := strings.TrimSpace(request.Caption)

		allowed, err := editCaption(r.Context(), config.CaptionEdit, r.PathValue("jumpID"), profile.Player.ID, caption, config.Now())
		if errors.Is(err, ErrJumpNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found.")
			return
		}
		if errors.Is(err, ErrAuthorGracePeriodExpired) {
			writeAPIError(w, http.StatusForbidden, "grace_period_expired", "Author Grace Period has expired.")
			return
		}
		if errors.Is(err, ErrInvalidCaption) {
			writeAPIError(w, http.StatusBadRequest, "invalid_caption", "Caption is required.")
			return
		}
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not update caption. Please try again.")
			return
		}
		if !allowed {
			writeAPIError(w, http.StatusForbidden, "not_performer", "Only the performer may edit the Caption.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/retract", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		allowed, err := retractJump(r.Context(), config.JumpRetract, r.PathValue("jumpID"), profile.Player.ID, config.Now())
		if errors.Is(err, ErrJumpNotFound) {
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found.")
			return
		}
		if errors.Is(err, ErrAuthorGracePeriodExpired) {
			writeAPIError(w, http.StatusForbidden, "grace_period_expired", "Author Grace Period has expired.")
			return
		}
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not retract Jump. Please try again.")
			return
		}
		if !allowed {
			writeAPIError(w, http.StatusForbidden, "not_performer", "Only the performer may retract the Jump.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "retracted"})
	})
	mux.HandleFunc("POST /v1/guest-sessions", func(w http.ResponseWriter, r *http.Request) {
		id, err := randomToken("guest_session")
		if err != nil {
			http.Error(w, "create guest session", http.StatusInternalServerError)
			return
		}
		if err := config.Judgment.CreateGuestSession(r.Context(), id); err != nil {
			http.Error(w, "create guest session", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id})
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/judgment", func(w http.ResponseWriter, r *http.Request) {
		var playerID, guestSessionID string
		authOK := false

		// Try authenticated path only if bearer token is present
		if r.Header.Get("Authorization") != "" {
			var profile MeResponse
			profile, authOK = signedInProfile(w, r, config)
			if authOK {
				playerID = profile.Player.ID
			}
		}

		var request struct {
			GuestSessionID *string `json:"guestSessionId"`
			Commitment     *int    `json:"commitment"`
			Transgression  *int    `json:"transgression"`
			Creativity     *int    `json:"creativity"`
			Presentation   *int    `json:"presentation"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if request.Commitment == nil || request.Transgression == nil || request.Creativity == nil || request.Presentation == nil {
			http.Error(w, "commitment, transgression, creativity, and presentation are required", http.StatusBadRequest)
			return
		}

		if request.GuestSessionID != nil && *request.GuestSessionID != "" {
			guestSessionID = *request.GuestSessionID
		}

		// Must have exactly one identity
		if !authOK && guestSessionID == "" {
			http.Error(w, "Authentication or guestSessionId required", http.StatusUnauthorized)
			return
		}
		if authOK && guestSessionID != "" {
			http.Error(w, "Cannot provide both authentication and guestSessionId", http.StatusBadRequest)
			return
		}

		judgment, ok, created, err := submitJudgment(
			r.Context(),
			config.Judgment,
			playerID,
			guestSessionID,
			"public",
			r.PathValue("jumpID"),
			*request.Commitment,
			*request.Transgression,
			*request.Creativity,
			*request.Presentation,
			config.Now(),
		)
		if errors.Is(err, ErrJumpNotFound) {
			http.Error(w, "Performed Jump not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrJudgingWindowClosed) {
			http.Error(w, "Judging Window closed", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrInvalidJudgmentScore) {
			http.Error(w, "Judgment scores must be between 1 and 4", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrAuthorGracePeriodActive) {
			http.Error(w, "Author Grace Period is still active", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrGuestCapReached) {
			http.Error(w, "Guest Judgment cap reached", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrInvalidJudgeIdentity) {
			http.Error(w, "Invalid judge identity", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "submit Judgment", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Judge required", http.StatusForbidden)
			return
		}

		status := http.StatusOK
		if created {
			status = http.StatusCreated
		}
		writeJSON(w, status, judgment)
	})
	mux.HandleFunc("POST /v1/opens/{year}/{month}/compute", func(w http.ResponseWriter, r *http.Request) {
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		yearStr := r.PathValue("year")
		monthStr := r.PathValue("month")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "invalid year", http.StatusBadRequest)
			return
		}
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			http.Error(w, "invalid month", http.StatusBadRequest)
			return
		}

		result := game.ComputeOpenScores(r.Context(), config.Open, game.ComputeOpenScoresInput{
			Year:  year,
			Month: month,
		}, config.Now())

		if errors.Is(result.Err, game.ErrOpenMonthNotClosed) {
			http.Error(w, "Open month has not soft-closed yet", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "compute Open scores", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			http.Error(w, "compute Open scores not allowed", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "computed"})
	})

	// GET /v1/feed — public Feed with optional auth
	mux.HandleFunc("GET /v1/feed", func(w http.ResponseWriter, r *http.Request) {
		viewer := optionalProfile(r, config)
		cursorStr := r.URL.Query().Get("cursor")
		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
				limit = parsed
			} else {
				http.Error(w, "invalid limit parameter", http.StatusBadRequest)
				return
			}
		}

		response, err := loadPublicFeed(r.Context(), config.PublicRead, config.Judgment, viewer, cursorStr, limit, config.Now())
		if errors.Is(err, ErrInvalidFeedCursor) {
			http.Error(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not load jumps. Please try again.")
			return
		}

		writeJSON(w, http.StatusOK, response)
	})

	// GET /v1/jumps/{jumpID} — public, unauthenticated Jump Detail
	mux.HandleFunc("GET /v1/jumps/{jumpID}", func(w http.ResponseWriter, r *http.Request) {
		viewer := optionalProfile(r, config)
		jumpID := r.PathValue("jumpID")

		response, found, err := loadPublicJumpDetail(r.Context(), config.PublicRead, config.Judgment, viewer, jumpID, config.Now())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not load jump detail. Please try again.")
			return
		}
		if !found {
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found. It may have been removed.")
			return
		}

		if response.Tombstone != nil {
			writeJSON(w, http.StatusOK, response.Tombstone)
			return
		}

		writeJSON(w, http.StatusOK, response.Detail)
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
