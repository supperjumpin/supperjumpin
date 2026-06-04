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

type StaticAuthVerifier map[string]AuthIdentity

func (v StaticAuthVerifier) Verify(token string) (AuthIdentity, bool) {
	identity, ok := v[token]
	return identity, ok
}

type ServerConfig struct {
	Auth         AuthVerifier
	Store        Store
	Now          func() time.Time
	JumpPlanning JumpPlanningFlow
	Judgment     JudgmentFlow
	PublicRead   PublicReadFlow
	Open         OpenFlow
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

		var cursorTS *time.Time
		var cursorID string
		if cursorStr != "" {
			ts, id, err := decodeCursor(cursorStr)
			if err != nil {
				http.Error(w, "invalid cursor", http.StatusBadRequest)
				return
			}
			cursorTS = &ts
			cursorID = id
		}

		// Fetch limit+1 to detect whether there's a next page
		cards, err := config.PublicRead.FeedJumps(r.Context(), cursorTS, cursorID, limit+1)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "Could not load jumps. Please try again.")
			return
		}

		// Attach viewer context to each card — batch already-judged check
		if viewer != nil {
			cardIDs := make([]string, len(cards))
			for i, c := range cards {
				cardIDs[i] = c.ID
			}
			judged, err := config.Judgment.HasJudgedJumps(r.Context(), viewer.Player.ID, cardIDs)
			if err != nil {
				http.Error(w, "Could not load judgment state", http.StatusInternalServerError)
				return
			}
			for i := range cards {
				hint := game.JudgmentEligibility(game.JumpSnapshot{
					ID:                   cards[i].ID,
					PlayerID:             cards[i].PerformerID,
					GracePeriodExpiresAt: cards[i].GracePeriodExpiresAt,
				}, viewer.Player.ID, judged[cards[i].ID], config.Now())
				cards[i].ViewerContext = viewerContextFromHint(hint)
			}
		}

		var nextCursor *string
		if len(cards) > limit {
			cards = cards[:limit]
			last := cards[len(cards)-1]
			c := encodeCursor(last.CreatedAt, last.ID)
			nextCursor = &c
		}

		writeJSON(w, http.StatusOK, PublicFeedResponse{Jumps: cards, NextCursor: nextCursor})
	})

	// GET /v1/jumps/{jumpID} — public, unauthenticated Jump Detail
	mux.HandleFunc("GET /v1/jumps/{jumpID}", func(w http.ResponseWriter, r *http.Request) {
		viewer := optionalProfile(r, config)
		jumpID := r.PathValue("jumpID")

		detail, found, err := config.PublicRead.JumpDetail(r.Context(), jumpID)
		if err != nil {
			http.Error(w, "Could not load jump detail", http.StatusInternalServerError)
			return
		}
		if !found {
			writeAPIError(w, http.StatusNotFound, "not_found", "Jump not found. It may have been removed.")
			return
		}

		// Tombstone for Removed Jumps
		if detail.Status == "Removed Jump" {
			removedAt := detail.CreatedAt
			if detail.RemovedAt != nil {
				removedAt = *detail.RemovedAt
			}
			writeJSON(w, http.StatusOK, JumpTombstone{
				ID:        detail.ID,
				Status:    "Removed Jump",
				Message:   "This Jump is no longer available",
				RemovedAt: removedAt.Format(time.RFC3339),
			})
			return
		}

		// Attach viewer context for all requests (unauthenticated = default canJudge: true)
		if viewer != nil {
			hasJudged, err := config.Judgment.HasJudgedJump(r.Context(), detail.ID, viewer.Player.ID)
			if err != nil {
				http.Error(w, "Could not load judgment state", http.StatusInternalServerError)
				return
			}
			hint := game.JudgmentEligibility(game.JumpSnapshot{
				ID:                   detail.ID,
				PlayerID:             detail.PerformerID,
				GracePeriodExpiresAt: detail.GracePeriodExpiresAt,
			}, viewer.Player.ID, hasJudged, config.Now())
			detail.ViewerContext = viewerContextFromHint(hint)
		} else {
			detail.ViewerContext = &ViewerContext{CanJudge: true}
		}

		writeJSON(w, http.StatusOK, detail)
	})

	return mux
}

// viewerContextFromHint converts a game.EligibilityHint into a transport-layer
// ViewerContext DTO.
func viewerContextFromHint(hint game.EligibilityHint) *ViewerContext {
	vc := &ViewerContext{CanJudge: hint.CanJudge}
	if hint.Reason != "" {
		r := hint.Reason
		vc.Reason = &r
	}
	if hint.GracePeriodEndsAt != nil {
		vc.GracePeriodEndsAt = hint.GracePeriodEndsAt
	}
	if !hint.CanJudge && hint.Reason == "already-judged" {
		vc.HasJudged = true
	}
	return vc
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
