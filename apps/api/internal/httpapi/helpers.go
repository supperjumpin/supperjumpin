package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/game"
)

func randomToken(kind string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate %s: %w", kind, err)
	}
	return kind + "_" + hex.EncodeToString(bytes), nil
}

func jumpFromGame(snap game.JumpSnapshot) Jump {
	return Jump{
		ID:                   snap.ID,
		PlayerID:             snap.PlayerID,
		Status:               snap.Status,
		Source:               snap.Source,
		Destination:          snap.Destination,
		Food:                 snap.Food,
		GracePeriodExpiresAt: snap.GracePeriodExpiresAt,
	}
}

func visiblePerformedStatus(status string) bool {
	return status == "Performed Jump" || status == "Judged Jump" || status == "Unjudged Jump" || status == "Disqualified Jump"
}

func stableID(kind string, value string) string {
	sum := sha256.Sum256([]byte(kind + ":" + value))
	return kind + "_" + hex.EncodeToString(sum[:])[:12]
}

func displayName(email string) string {
	name, _, ok := strings.Cut(email, "@")
	if !ok || name == "" {
		return "player"
	}
	return name
}

// optionalProfile attempts to extract the signed-in player profile,
// returning nil if no valid token is present (no 401 written).
func optionalProfile(r *http.Request, config ServerConfig) *MeResponse {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		return nil
	}
	identity, ok := config.Auth.Verify(token)
	if !ok {
		return nil
	}
	profile, err := config.Store.BootstrapIdentity(r.Context(), identity)
	if err != nil {
		return nil
	}
	AddRequestLogFields(r.Context(), slog.String("actor_type", "player"), slog.String("player_id", profile.Player.ID))
	return &profile
}

// feedCursor encodes a timestamp + ID pair for cursor-based pagination.
type feedCursor struct {
	CreatedAt int64  `json:"t"`
	LastID    string `json:"i"`
}

func encodeCursor(t time.Time, id string) string {
	fc := feedCursor{CreatedAt: t.UnixMilli(), LastID: id}
	data, _ := json.Marshal(fc)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(cursor string) (time.Time, string, error) {
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("decode cursor: %w", err)
	}
	if len(data) > 256 {
		return time.Time{}, "", fmt.Errorf("cursor too long")
	}
	var fc feedCursor
	if err := json.Unmarshal(data, &fc); err != nil {
		return time.Time{}, "", fmt.Errorf("unmarshal cursor: %w", err)
	}
	if fc.CreatedAt <= 0 || fc.LastID == "" {
		return time.Time{}, "", fmt.Errorf("invalid cursor values")
	}
	// Reject cursors more than 1 hour in the future to prevent cursor manipulation
	if fc.CreatedAt > time.Now().UnixMilli()+3600000 {
		return time.Time{}, "", fmt.Errorf("cursor timestamp too far in the future")
	}
	return time.UnixMilli(fc.CreatedAt), fc.LastID, nil
}
