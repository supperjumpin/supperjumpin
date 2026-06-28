package httpapi

import (
	"encoding/json"
	"errors"
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
	mux.HandleFunc("POST /v1/rounds/{roundId}/commits", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds/{roundId}/commits", "commit_to_round")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		roundID := r.PathValue("roundId")
		now := config.Now()

		result, err := game.CommitToRound(r.Context(), config.Store, game.CommitToRoundInput{
			RoundID:  roundID,
			PlayerID: profile.Player.ID,
		}, now)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "commit_to_round_failed", err)
			http.Error(w, "commit to round", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "commit_to_round_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		writeJSON(w, http.StatusCreated, CommitResponse{CommitID: result.CommitID})
	})
	mux.HandleFunc("POST /v1/rounds/{roundId}/jumps", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds/{roundId}/jumps", "submit_jump")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request SubmitJumpRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if request.Caption == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_caption", nil)
			http.Error(w, "caption is required", http.StatusBadRequest)
			return
		}
		if len(request.EvidenceURLs) == 0 {
			recordHTTPError(r, http.StatusBadRequest, "missing_evidence", nil)
			http.Error(w, "at least one evidence url is required", http.StatusBadRequest)
			return
		}

		roundID := r.PathValue("roundId")
		now := config.Now()

		result, err := game.SubmitJump(r.Context(), config.Store, game.SubmitJumpInput{
			RoundID:      roundID,
			PlayerID:     profile.Player.ID,
			Caption:      request.Caption,
			EvidenceURLs: request.EvidenceURLs,
		}, now)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "submit_jump_failed", err)
			http.Error(w, "submit jump", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "submit_jump_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		jumpDTO := JumpDTO{
			ID:          result.Jump.ID,
			RoundID:     result.Jump.RoundID,
			PlayerID:    result.Jump.PlayerID,
			Caption:     result.Jump.Caption,
			EvidenceURLs: result.Jump.EvidenceURLs,
			SubmittedAt: result.Jump.SubmittedAt.Format(time.RFC3339),
		}

		writeJSON(w, http.StatusCreated, SubmitJumpResponse{Jump: jumpDTO})
	})
	mux.HandleFunc("GET /v1/rounds/{roundId}/jumps", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/rounds/{roundId}/jumps", "list_round_jumps")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		roundID := r.PathValue("roundId")
		result, err := game.ListJumpsForRound(r.Context(), config.Store, roundID, profile.Player.ID)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "list_round_jumps_failed", err)
			http.Error(w, "list round jumps", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "list_round_jumps_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		jumps := make([]JumpDTO, 0, len(result.Jumps))
		for _, j := range result.Jumps {
			dto := JumpDTO{
				ID:                 j.ID,
				RoundID:            j.RoundID,
				PlayerID:           j.PlayerID,
				SealedViewer:       j.SealedViewer,
				PlayerHasCommitted:  j.PlayerHasCommitted,
				PlayerHasSubmitted:  j.PlayerHasSubmitted,
			}
			if j.SubmittedAt.Unix() > 0 {
				dto.SubmittedAt = j.SubmittedAt.Format(time.RFC3339)
			}
			if !j.SealedViewer {
				dto.Caption = j.Caption
				dto.EvidenceURLs = j.EvidenceURLs
			}
			jumps = append(jumps, dto)
		}

		// Get round status for counts
		status, err := config.Store.GetRoundStatus(r.Context(), roundID)
		if err != nil && !errors.Is(err, game.ErrRoundNotFound) {
			recordHTTPError(r, http.StatusInternalServerError, "get_round_status_failed", err)
			http.Error(w, "get round status", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, ListJumpsResponse{
			Jumps:          jumps,
			CommitCount:    status.CommitCount,
			SubmissionCount: status.SubmissionCount,
		})
	})
	mux.HandleFunc("POST /v1/rounds/{roundId}/reveal", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds/{roundId}/reveal", "evaluate_reveal")
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		roundID := r.PathValue("roundId")
		now := config.Now()

		result, err := game.EvaluateReveal(r.Context(), config.Store, roundID, now)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "evaluate_reveal_failed", err)
			http.Error(w, "evaluate reveal", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "evaluate_reveal_forbidden", result.Err)
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

		writeJSON(w, http.StatusOK, RevealRoundResponse{Round: round, Revealed: result.Revealed})
	})
	mux.HandleFunc("GET /v1/rounds/{roundId}/recap", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/rounds/{roundId}/recap", "get_round_recap")
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		roundID := r.PathValue("roundId")
		AddRequestLogField(r.Context(), "round_id", roundID)

		result, err := game.AssembleRecap(r.Context(), config.Store, roundID)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "get_round_recap_failed", err)
			http.Error(w, "get round recap", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "get_round_recap_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		recap := result.Recap
		AddRequestLogField(r.Context(), "community_id", recap.CommunityID)

		jumps := make([]RecapJumpEntryDTO, 0, len(recap.Jumps))
		for _, j := range recap.Jumps {
			jumps = append(jumps, RecapJumpEntryDTO{
				JumpID:       j.JumpID,
				PlayerID:     j.PlayerID,
				Caption:      j.Caption,
				EvidenceURLs: j.EvidenceURLs,
				SubmittedAt:  j.SubmittedAt.Format(time.RFC3339),
				StampCounts:  j.StampCounts,
				TotalStamps:  j.TotalStamps,
			})
		}

		comments := make([]CommentDTO, 0, len(recap.Comments))
		for _, c := range recap.Comments {
			comments = append(comments, CommentDTO{
				ID:        c.ID,
				RoundID:   c.RoundID,
				JumpID:    c.JumpID,
				PlayerID:  c.PlayerID,
				Body:      c.Body,
				CreatedAt: c.CreatedAt.Format(time.RFC3339),
			})
		}

		ghostJumpers := make([]GhostJumperDTO, 0, len(recap.GhostJumpers))
		for _, g := range recap.GhostJumpers {
			ghostJumpers = append(ghostJumpers, GhostJumperDTO{
				PlayerID:    g.PlayerID,
				CommittedAt: g.CommittedAt.Format(time.RFC3339),
			})
		}

		lore := make([]LoreEntryDTO, 0, len(recap.Lore))
		for _, e := range recap.Lore {
			lore = append(lore, LoreEntryDTO{
				JumpID:       e.JumpID,
				RoundID:      e.RoundID,
				JumpCaption:  e.JumpCaption,
				JumpPlayerID: e.JumpPlayerID,
				StampCounts:  e.StampCounts,
				TotalStamps:  e.TotalStamps,
			})
		}

standoutStamps := make([]StandoutStampDTO, 0, len(recap.StandoutStamps))
		for _, s := range recap.StandoutStamps {
			standoutStamps = append(standoutStamps, StandoutStampDTO{
				JumpID: s.JumpID,
				Stance: s.Stance,
				Count:  s.Count,
			})
		}

		standoutComments := make([]CommentDTO, 0, len(recap.StandoutComments))
		for _, c := range recap.StandoutComments {
			standoutComments = append(standoutComments, CommentDTO{
				ID:        c.ID,
				RoundID:   c.RoundID,
				JumpID:    c.JumpID,
				PlayerID:  c.PlayerID,
				Body:      c.Body,
				CreatedAt: c.CreatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, RecapResponse{
			RoundID:       recap.RoundID,
			CommunityID:   recap.CommunityID,
			PromptID:      recap.PromptID,
			Status:        recap.Status,
			RevealBy:      recap.RevealBy.Format(time.RFC3339),
			CreatedBy:     recap.CreatedBy,
			CreatedAt:     recap.CreatedAt.Format(time.RFC3339),
			Jumps:         jumps,
			Comments:      comments,
			GhostJumpers:  ghostJumpers,
			Lore:          lore,
			NextRoundHook: NextRoundHookDTO{
				ActiveRoundID: recap.NextRoundHook.ActiveRoundID,
				PromptID:      recap.NextRoundHook.PromptID,
			},
			StandoutStamps:   standoutStamps,
			StandoutComments: standoutComments,
		})
	})
	mux.HandleFunc("GET /v1/rounds/{roundId}/jumps/{jumpId}", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/rounds/{roundId}/jumps/{jumpId}", "get_jump")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		jumpID := r.PathValue("jumpId")
		result, err := game.GetJump(r.Context(), config.Store, jumpID, profile.Player.ID)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "get_jump_failed", err)
			http.Error(w, "get jump", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "get_jump_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		dto := JumpDTO{
			ID:                 result.Jump.ID,
			RoundID:            result.Jump.RoundID,
			PlayerID:           result.Jump.PlayerID,
			SealedViewer:       result.Jump.SealedViewer,
			PlayerHasCommitted:  result.Jump.PlayerHasCommitted,
			PlayerHasSubmitted:  result.Jump.PlayerHasSubmitted,
		}
		if result.Jump.SubmittedAt.Unix() > 0 {
			dto.SubmittedAt = result.Jump.SubmittedAt.Format(time.RFC3339)
		}
		if !result.Jump.SealedViewer {
			dto.Caption = result.Jump.Caption
			dto.EvidenceURLs = result.Jump.EvidenceURLs
		}

		writeJSON(w, http.StatusOK, GetJumpResponse{Jump: dto})
	})
	mux.HandleFunc("GET /v1/stamp-catalog", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/stamp-catalog", "list_stamp_catalog")
		AddRequestLogField(r.Context(), "actor_type", "public")

		result, err := game.ListStampCatalog(r.Context(), config.Store)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "list_stamp_catalog_failed", err)
			http.Error(w, "list stamp catalog", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusInternalServerError, "list_stamp_catalog_failed", result.Err)
			http.Error(w, "list stamp catalog", http.StatusInternalServerError)
			return
		}

		stamps := make([]StampDTO, 0, len(result.Stamps))
		for _, s := range result.Stamps {
			stamps = append(stamps, StampDTO{
				ID:     s.ID,
				Stance: s.Stance,
				Label:  s.Label,
				Glyph:  s.Glyph,
				Copy:   s.Copy,
			})
		}

		writeJSON(w, http.StatusOK, StampCatalogResponse{Stamps: stamps})
	})
	mux.HandleFunc("POST /v1/rounds/{roundId}/jumps/{jumpId}/reactions", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds/{roundId}/jumps/{jumpId}/reactions", "apply_reaction")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request ApplyReactionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if request.StampID == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_stamp_id", nil)
			http.Error(w, "stampId is required", http.StatusBadRequest)
			return
		}

		jumpID := r.PathValue("jumpId")
		now := config.Now()

		result, err := game.ApplyReaction(r.Context(), config.Store, game.ApplyReactionInput{
			JumpID:   jumpID,
			StampID:  request.StampID,
			PlayerID: profile.Player.ID,
		}, now)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "apply_reaction_failed", err)
			http.Error(w, "apply reaction", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "apply_reaction_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		reaction := ReactionDTO{
			ID:        result.Reaction.ID,
			StampID:   result.Reaction.StampID,
			JumpID:    result.Reaction.JumpID,
			PlayerID:  result.Reaction.PlayerID,
			CreatedAt: result.Reaction.CreatedAt.Format(time.RFC3339),
		}

		writeJSON(w, http.StatusCreated, ApplyReactionResponse{Reaction: reaction})
	})
	mux.HandleFunc("POST /v1/rounds/{roundId}/comments", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds/{roundId}/comments", "post_round_comment")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request PostCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(request.Body) == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_comment_body", nil)
			http.Error(w, "body is required", http.StatusBadRequest)
			return
		}

		roundID := r.PathValue("roundId")
		now := config.Now()

		result, err := game.PostComment(r.Context(), config.Store, game.PostCommentInput{
			RoundID:  roundID,
			JumpID:   "",
			PlayerID: profile.Player.ID,
			Body:     strings.TrimSpace(request.Body),
		}, now)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "post_round_comment_failed", err)
			http.Error(w, "post comment", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "post_round_comment_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		comment := CommentDTO{
			ID:        result.Comment.ID,
			RoundID:   result.Comment.RoundID,
			PlayerID:  result.Comment.PlayerID,
			Body:      result.Comment.Body,
			CreatedAt: result.Comment.CreatedAt.Format(time.RFC3339),
		}

		writeJSON(w, http.StatusCreated, PostCommentResponse{Comment: comment})
	})
	mux.HandleFunc("GET /v1/rounds/{roundId}/comments", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/rounds/{roundId}/comments", "list_round_comments")
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		roundID := r.PathValue("roundId")

		result, err := game.ListComments(r.Context(), config.Store, game.ListCommentsInput{
			RoundID: roundID,
			JumpID:  "",
		})
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "list_round_comments_failed", err)
			http.Error(w, "list comments", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "list_round_comments_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		comments := make([]CommentDTO, 0, len(result.Comments))
		for _, c := range result.Comments {
			comments = append(comments, CommentDTO{
				ID:        c.ID,
				RoundID:   c.RoundID,
				PlayerID:  c.PlayerID,
				Body:      c.Body,
				CreatedAt: c.CreatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, ListCommentsResponse{Comments: comments})
	})
	mux.HandleFunc("POST /v1/rounds/{roundId}/jumps/{jumpId}/comments", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "POST /v1/rounds/{roundId}/jumps/{jumpId}/comments", "post_jump_comment")
		profile, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		var request PostCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			recordHTTPError(r, http.StatusBadRequest, "invalid_json", nil)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(request.Body) == "" {
			recordHTTPError(r, http.StatusBadRequest, "missing_comment_body", nil)
			http.Error(w, "body is required", http.StatusBadRequest)
			return
		}

		roundID := r.PathValue("roundId")
		jumpID := r.PathValue("jumpId")
		now := config.Now()

		result, err := game.PostComment(r.Context(), config.Store, game.PostCommentInput{
			RoundID:  roundID,
			JumpID:   jumpID,
			PlayerID: profile.Player.ID,
			Body:     strings.TrimSpace(request.Body),
		}, now)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "post_jump_comment_failed", err)
			http.Error(w, "post comment", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "post_jump_comment_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		comment := CommentDTO{
			ID:        result.Comment.ID,
			RoundID:   result.Comment.RoundID,
			JumpID:    result.Comment.JumpID,
			PlayerID:  result.Comment.PlayerID,
			Body:      result.Comment.Body,
			CreatedAt: result.Comment.CreatedAt.Format(time.RFC3339),
		}

		writeJSON(w, http.StatusCreated, PostCommentResponse{Comment: comment})
	})
	mux.HandleFunc("GET /v1/rounds/{roundId}/jumps/{jumpId}/comments", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/rounds/{roundId}/jumps/{jumpId}/comments", "list_jump_comments")
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		roundID := r.PathValue("roundId")
		jumpID := r.PathValue("jumpId")

		result, err := game.ListComments(r.Context(), config.Store, game.ListCommentsInput{
			RoundID: roundID,
			JumpID:  jumpID,
		})
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "list_jump_comments_failed", err)
			http.Error(w, "list comments", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "list_jump_comments_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		comments := make([]CommentDTO, 0, len(result.Comments))
		for _, c := range result.Comments {
			comments = append(comments, CommentDTO{
				ID:        c.ID,
				RoundID:   c.RoundID,
				JumpID:    c.JumpID,
				PlayerID:  c.PlayerID,
				Body:      c.Body,
				CreatedAt: c.CreatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, ListCommentsResponse{Comments: comments})
	})
	mux.HandleFunc("GET /v1/communities/{communityId}/lore", func(w http.ResponseWriter, r *http.Request) {
		setRequestOperation(r, "GET /v1/communities/{communityId}/lore", "get_community_lore")
		_, ok := signedInProfile(w, r, config)
		if !ok {
			return
		}

		communityID := r.PathValue("communityId")
		AddRequestLogField(r.Context(), "community_id", communityID)

		result, err := game.DeriveCommunityLore(r.Context(), config.Store, communityID)
		if err != nil {
			recordHTTPError(r, http.StatusInternalServerError, "get_community_lore_failed", err)
			http.Error(w, "get community lore", http.StatusInternalServerError)
			return
		}
		if !result.Allowed {
			recordHTTPError(r, http.StatusForbidden, "get_community_lore_forbidden", result.Err)
			http.Error(w, result.Err.Error(), http.StatusForbidden)
			return
		}

		entries := make([]LoreEntryDTO, 0, len(result.Entries))
		for _, e := range result.Entries {
			entries = append(entries, LoreEntryDTO{
				JumpID:       e.JumpID,
				RoundID:      e.RoundID,
				JumpCaption:  e.JumpCaption,
				JumpPlayerID: e.JumpPlayerID,
				StampCounts:  e.StampCounts,
				TotalStamps:  e.TotalStamps,
			})
		}

		writeJSON(w, http.StatusOK, CommunityLoreResponse{Entries: entries})
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
