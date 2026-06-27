package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
)

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
