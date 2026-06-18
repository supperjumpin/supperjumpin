package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
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
	Logger       *slog.Logger
	GuestCap     int
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
	mux.HandleFunc("POST /v1/jumps", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/jumps", "create_performed_jump")
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
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		source := strings.TrimSpace(request.Source)
		destination := strings.TrimSpace(request.Destination)
		food := strings.TrimSpace(request.Food)
		caption := strings.TrimSpace(request.Caption)
		mediaObjectKey := strings.TrimSpace(request.MediaObjectKey)
		if source == "" || destination == "" || food == "" || caption == "" || mediaObjectKey == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_jump_fields", nil)
			http.Error(w, "Source, Destination, Food, Caption, and mediaObjectKey are required", http.StatusBadRequest)
			return
		}

		jump, err := createPerformedJump(r.Context(), config.JumpPlanning, profile.Player, source, destination, food, caption, mediaObjectKey, config.Now())
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "create_performed_jump_failed", err)
			http.Error(w, "create Performed Jump", http.StatusInternalServerError)
			return
		}
		AddRequestLogField(r.Context(), "jump_id", jump.ID)

		writeJSON(w, http.StatusCreated, jump)
	})
	mux.HandleFunc("PATCH /v1/jumps/{jumpID}", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "PATCH /v1/jumps/{jumpID}", "edit_jump_caption")
		AddRequestLogField(r.Context(), "jump_id", r.PathValue("jumpID"))
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Caption string `json:"caption"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		caption := strings.TrimSpace(request.Caption)

		allowed, err := editCaption(r.Context(), config.CaptionEdit, r.PathValue("jumpID"), profile.Player.ID, caption, config.Now())
		if errors.Is(err, ErrJumpNotFound) {
			recordHTTPError(r, http.StatusNotFound, "not_found", nil)
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found.")
			return
		}
		if errors.Is(err, ErrAuthorGracePeriodExpired) {
			recordHTTPError(r, http.StatusForbidden, "grace_period_expired", nil)
			writeAPIError(w, http.StatusForbidden, "grace_period_expired", "Author Grace Period has expired.")
			return
		}
		if errors.Is(err, ErrInvalidCaption) {
			recordHTTPError(r, http.StatusBadRequest, "invalid_caption", nil)
			writeAPIError(w, http.StatusBadRequest, "invalid_caption", "Caption is required.")
			return
		}
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "internal_error", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not update caption. Please try again.")
			return
		}
		if !allowed {
			recordHTTPError(r, http.StatusForbidden, "not_performer", nil)
			writeAPIError(w, http.StatusForbidden, "not_performer", "Only the performer may edit the Caption.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/retract", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/jumps/{jumpID}/retract", "retract_jump")
		AddRequestLogField(r.Context(), "jump_id", r.PathValue("jumpID"))
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		allowed, err := retractJump(r.Context(), config.JumpRetract, r.PathValue("jumpID"), profile.Player.ID, config.Now())
		if errors.Is(err, ErrJumpNotFound) {
			recordHTTPError(r, http.StatusNotFound, "not_found", nil)
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found.")
			return
		}
		if errors.Is(err, ErrAuthorGracePeriodExpired) {
			recordHTTPError(r, http.StatusForbidden, "grace_period_expired", nil)
			writeAPIError(w, http.StatusForbidden, "grace_period_expired", "Author Grace Period has expired.")
			return
		}
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "internal_error", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not retract Jump. Please try again.")
			return
		}
		if !allowed {
			recordHTTPError(r, http.StatusForbidden, "not_performer", nil)
			writeAPIError(w, http.StatusForbidden, "not_performer", "Only the performer may retract the Jump.")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "retracted"})
	})
	mux.HandleFunc("POST /v1/guest-sessions", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/guest-sessions", "create_guest_session")
		AddRequestLogField(r.Context(), "actor_type", "guest")
		id, err := randomToken("guest_session")
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "create_guest_session_id_failed", err)
			http.Error(w, "create guest session", http.StatusInternalServerError)
			return
		}
		if err := config.Judgment.CreateGuestSession(r.Context(), id); err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "create_guest_session_failed", err)
			http.Error(w, "create guest session", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": id})
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/judgment", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/jumps/{jumpID}/judgment", "submit_judgment")
		AddRequestLogField(r.Context(), "jump_id", r.PathValue("jumpID"))
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
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if request.GuestSessionID != nil && *request.GuestSessionID != "" {
			guestSessionID = *request.GuestSessionID
			if !authOK {
				AddRequestLogField(r.Context(), "actor_type", "guest")
			}
		}
		if request.Commitment == nil || request.Transgression == nil || request.Creativity == nil || request.Presentation == nil {
			recordHTTPError(r, http.StatusBadRequest, "missing_judgment_scores", nil)
			http.Error(w, "commitment, transgression, creativity, and presentation are required", http.StatusBadRequest)
			return
		}

		// Must have exactly one identity
		if !authOK && guestSessionID == "" {
			recordHTTPError(r, http.StatusUnauthorized, "missing_judge_identity", nil)
			http.Error(w, "Authentication or guestSessionId required", http.StatusUnauthorized)
			return
		}
		if authOK && guestSessionID != "" {
			recordHTTPError(r, http.StatusBadRequest, "multiple_judge_identities", nil)
			http.Error(w, "Cannot provide both authentication and guestSessionId", http.StatusBadRequest)
			return
		}

		judgment, ok, err := submitJudgment(
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
			config.GuestCap,
		)
		if errors.Is(err, ErrJumpNotFound) {
			recordHTTPError(r, http.StatusNotFound, "not_found", nil)
			writeAPIError(w, http.StatusNotFound, "not_found", "Performed Jump not found.")
			return
		}
		if errors.Is(err, ErrJudgingWindowClosed) {
			recordHTTPError(r, http.StatusConflict, "window_closed", nil)
			writeAPIError(w, http.StatusConflict, "window_closed", "Judging Window closed.")
			return
		}
		if errors.Is(err, ErrInvalidJudgmentScore) {
			recordHTTPError(r, http.StatusBadRequest, "invalid_score", nil)
			writeAPIError(w, http.StatusBadRequest, "invalid_score", "Judgment scores must be between 1 and 4.")
			return
		}
		if errors.Is(err, ErrAuthorGracePeriodActive) {
			recordHTTPError(r, http.StatusForbidden, "grace_period", nil)
			writeAPIError(w, http.StatusForbidden, "grace_period", "Author Grace Period is still active.")
			return
		}
		if errors.Is(err, ErrGuestCapReached) {
			recordHTTPError(r, http.StatusForbidden, "guest_cap", nil)
			writeAPIError(w, http.StatusForbidden, "guest_cap", "Guest Judgment cap reached.")
			return
		}
		if errors.Is(err, ErrInvalidJudgeIdentity) {
			recordHTTPError(r, http.StatusBadRequest, "invalid_judge_identity", nil)
			writeAPIError(w, http.StatusBadRequest, "invalid_judge_identity", "Invalid judge identity.")
			return
		}
		if errors.Is(err, ErrAlreadyJudged) {
			recordHTTPError(r, http.StatusConflict, "already_judged", nil)
			writeAPIError(w, http.StatusConflict, "already_judged", "Judge has already submitted a Judgment for this Jump.")
			return
		}
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "internal_error", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not submit Judgment. Please try again.")
			return
		}
		if !ok {
			recordHTTPError(r, http.StatusForbidden, "self_judging", nil)
			writeAPIError(w, http.StatusForbidden, "self_judging", "Judge must be a different Player than the performer.")
			return
		}
		AddRequestLogField(r.Context(), "judgment_id", judgment.ID)

		writeJSON(w, http.StatusCreated, judgment)
	})
	mux.HandleFunc("POST /v1/opens/{year}/{month}/compute", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/opens/{year}/{month}/compute", "compute_open_scores")
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		yearStr := r.PathValue("year")
		monthStr := r.PathValue("month")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_year", nil)
			http.Error(w, "invalid year", http.StatusBadRequest)
			return
		}
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			recordHTTPError(r, http.StatusBadRequest, "invalid_month", nil)
			http.Error(w, "invalid month", http.StatusBadRequest)
			return
		}
		AddRequestLogFields(r.Context(), slog.Int("open_year", year), slog.Int("open_month", month))

		result := game.ComputeOpenScores(r.Context(), config.Open, game.ComputeOpenScoresInput{
			Year:  year,
			Month: month,
		}, config.Now())

		if errors.Is(result.Err, game.ErrOpenMonthNotClosed) {
			recordHTTPError(r, http.StatusConflict, "open_month_not_closed", nil)
			http.Error(w, "Open month has not soft-closed yet", http.StatusConflict)
			return
		}
		if result.Err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "compute_open_scores_failed", result.Err)
			http.Error(w, "compute Open scores", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "compute_open_scores_not_allowed", nil)
			http.Error(w, "compute Open scores not allowed", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "computed"})
	})

	// GET /v1/feed — public Feed with optional auth
	mux.HandleFunc("GET /v1/feed", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/feed", "load_feed")
		viewer := optionalProfile(r, config)
		if viewer == nil {
			AddRequestLogField(r.Context(), "actor_type", "public")
		}
		cursorStr := r.URL.Query().Get("cursor")
		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
				limit = parsed
			} else {
				recordHTTPError(r, http.StatusBadRequest, "invalid_limit", nil)
				http.Error(w, "invalid limit parameter", http.StatusBadRequest)
				return
			}
		}

		response, err := loadPublicFeed(r.Context(), config.PublicRead, config.Judgment, viewer, cursorStr, limit, config.Now())
		if errors.Is(err, ErrInvalidFeedCursor) {
			recordHTTPError(r, http.StatusBadRequest, "invalid_cursor", nil)
			http.Error(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "internal_error", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not load jumps. Please try again.")
			return
		}

		writeJSON(w, http.StatusOK, response)
	})

	// GET /v1/jumps/{jumpID} — public, unauthenticated Jump Detail
	mux.HandleFunc("GET /v1/jumps/{jumpID}", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/jumps/{jumpID}", "load_jump_detail")
		viewer := optionalProfile(r, config)
		if viewer == nil {
			AddRequestLogField(r.Context(), "actor_type", "public")
		}
		jumpID := r.PathValue("jumpID")
		AddRequestLogField(r.Context(), "jump_id", jumpID)

		response, found, err := loadPublicJumpDetail(r.Context(), config.PublicRead, config.Judgment, viewer, jumpID, config.Now())
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "internal_error", err)
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not load jump detail. Please try again.")
			return
		}
		if !found {
			recordHTTPError(r, http.StatusNotFound, "not_found", nil)
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found. It may have been removed.")
			return
		}

		if response.Tombstone != nil {
			writeJSON(w, http.StatusOK, response.Tombstone)
			return
		}

		writeJSON(w, http.StatusOK, response.Detail)
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
