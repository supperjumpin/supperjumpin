package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
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
	DB    Persistence
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

		home, err := createGroup(r.Context(), config.DB, profile.Player, name)
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

		groups, err := listGroups(r.Context(), config.DB, profile.Player)
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

		home, ok, err := groupHomeHandler(r.Context(), config.DB, profile.Player, r.PathValue("groupID"))
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

		invite, ok, err := createInvite(r.Context(), config.DB, profile.Player, r.PathValue("groupID"))
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

		home, status, err := acceptInvite(r.Context(), config.DB, profile.Player, r.PathValue("token"))
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
		case InviteMember:
			http.Error(w, "Player already has a Group Membership", http.StatusConflict)
			return
		}

		writeJSON(w, http.StatusOK, home)
	})
	mux.HandleFunc("POST /v1/groups/{groupID}/seasons", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var body struct {
			SubmissionDeadline string `json:"submissionDeadline"`
			JudgingDeadline    string `json:"judgingDeadline"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		submissionDeadline, err := time.Parse(time.RFC3339, body.SubmissionDeadline)
		if err != nil {
			http.Error(w, "submissionDeadline must be ISO 8601 format", http.StatusBadRequest)
			return
		}
		judgingDeadline, err := time.Parse(time.RFC3339, body.JudgingDeadline)
		if err != nil {
			http.Error(w, "judgingDeadline must be ISO 8601 format", http.StatusBadRequest)
			return
		}

		home, ok, err := startSeason(r.Context(), config.DB, profile.Player, r.PathValue("groupID"), submissionDeadline, judgingDeadline)
		if errors.Is(err, ErrSeasonAlreadyOpen) {
			http.Error(w, "Group already has an active or closing Season", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "start Season", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, home)
	})
	mux.HandleFunc("POST /v1/seasons/{seasonID}/close-submissions", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		home, ok, err := closeSeasonSubmissions(r.Context(), config.DB, profile.Player, r.PathValue("seasonID"))
		if errors.Is(err, ErrSeasonNotFound) {
			http.Error(w, "Season not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "close Season submissions", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Season Commissioner or Group Admin required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, home)
	})
	mux.HandleFunc("POST /v1/seasons/{seasonID}/finalize", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		home, ok, err := finalizeSeason(r.Context(), config.DB, profile.Player, r.PathValue("seasonID"))
		if errors.Is(err, ErrSeasonNotFound) {
			http.Error(w, "Season not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "finalize Season", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Season Commissioner or Group Admin required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, home)
	})
	mux.HandleFunc("GET /v1/seasons/{seasonID}/history", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		history, ok, err := seasonHistory(r.Context(), config.DB, profile.Player, r.PathValue("seasonID"))
		if errors.Is(err, ErrSeasonNotFound) {
			http.Error(w, "Season not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "get Season history", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, history)
	})
	mux.HandleFunc("POST /v1/groups/{groupID}/ideas", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Food        string `json:"food"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		source := strings.TrimSpace(request.Source)
		destination := strings.TrimSpace(request.Destination)
		food := strings.TrimSpace(request.Food)
		if source == "" || destination == "" || food == "" {
			http.Error(w, "Source, Destination, and Food are required", http.StatusBadRequest)
			return
		}

		idea, ok, err := createIdea(r.Context(), config.DB, profile.Player, r.PathValue("groupID"), source, destination, food)
		if err != nil {
			http.Error(w, "create Idea", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, idea)
	})
	mux.HandleFunc("POST /v1/ideas/{ideaID}/planned-jump", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			OffSeason bool `json:"offSeason"`
		}
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
		}
		planned, ok, err := createPlannedJump(r.Context(), config.DB, profile.Player, r.PathValue("ideaID"), request.OffSeason)
		if errors.Is(err, ErrJumpNotFound) {
			http.Error(w, "Idea not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "create Planned Jump", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, planned)
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/evidence-upload-authorizations", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			ContentType string `json:"contentType"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		contentType := strings.TrimSpace(request.ContentType)
		if contentType == "" {
			http.Error(w, "contentType is required", http.StatusBadRequest)
			return
		}

		authorization, ok, err := authorizeEvidenceUpload(r.Context(), config.DB, profile.Player, r.PathValue("jumpID"), contentType)
		if errors.Is(err, ErrJumpNotFound) {
			http.Error(w, "Planned Jump not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "authorize Evidence upload", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "performer required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, authorization)
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/evidence", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			UploadAuthorizationID string `json:"uploadAuthorizationId"`
			Caption               string `json:"caption"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		uploadAuthorizationID := strings.TrimSpace(request.UploadAuthorizationID)
		caption := strings.TrimSpace(request.Caption)
		if uploadAuthorizationID == "" || caption == "" {
			http.Error(w, "uploadAuthorizationId and caption are required", http.StatusBadRequest)
			return
		}

		submission, ok, err := submitEvidence(r.Context(), config.DB, profile.Player, r.PathValue("jumpID"), uploadAuthorizationID, caption)
		if errors.Is(err, ErrJumpNotFound) {
			http.Error(w, "Planned Jump not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrEvidenceUploadAuthorizationNotFound) {
			http.Error(w, "Evidence upload authorization not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrSubmissionWindowClosed) {
			http.Error(w, "Submission Window closed", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "submit Evidence", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "performer required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, submission)
	})
	mux.HandleFunc("POST /v1/jumps/{jumpID}/judgment", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Difficulty    *int `json:"difficulty"`
			Transgression *int `json:"transgression"`
			Creativity    *int `json:"creativity"`
			Presentation  *int `json:"presentation"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if request.Difficulty == nil || request.Transgression == nil || request.Creativity == nil || request.Presentation == nil {
			http.Error(w, "difficulty, transgression, creativity, and presentation are required", http.StatusBadRequest)
			return
		}

		judgment, ok, created, err := submitJudgment(
			r.Context(),
			config.DB,
			profile.Player,
			r.PathValue("jumpID"),
			*request.Difficulty,
			*request.Transgression,
			*request.Creativity,
			*request.Presentation,
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
			http.Error(w, "Judgment scores must be between 0 and 10", http.StatusBadRequest)
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
	mux.HandleFunc("POST /v1/jumps/{jumpID}/disputes", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Concern string `json:"concern"`
			Details string `json:"details"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		concern := strings.TrimSpace(request.Concern)
		details := strings.TrimSpace(request.Details)
		if concern == "" || details == "" {
			http.Error(w, "concern and details are required", http.StatusBadRequest)
			return
		}

		dispute, ok, err := createDispute(r.Context(), config.DB, profile.Player, r.PathValue("jumpID"), concern, details)
		if errors.Is(err, ErrJumpNotFound) {
			http.Error(w, "Visible Jump not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvalidDisputeConcern) {
			http.Error(w, "Dispute concern must be House Rules, Credibility, Source, Destination, Food, duplicate, or other", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "create Dispute", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "Group Membership required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, dispute)
	})
	mux.HandleFunc("POST /v1/disputes/{disputeID}/resolution", func(w http.ResponseWriter, r *http.Request) {
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request struct {
			Resolution       string `json:"resolution"`
			ResolutionReason string `json:"resolutionReason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		resolution := strings.TrimSpace(request.Resolution)
		resolutionReason := strings.TrimSpace(request.ResolutionReason)
		if resolution == "" || resolutionReason == "" {
			http.Error(w, "resolution and resolutionReason are required", http.StatusBadRequest)
			return
		}

		resolved, ok, err := resolveDispute(r.Context(), config.DB, profile.Player, r.PathValue("disputeID"), resolution, resolutionReason)
		if errors.Is(err, ErrDisputeNotFound) {
			http.Error(w, "Dispute not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvalidDisputeResolution) {
			http.Error(w, "Dispute resolution must be No Action, Disqualified Jump, or Removed Jump", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "resolve Dispute", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "authorized resolver required", http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusOK, resolved)
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
