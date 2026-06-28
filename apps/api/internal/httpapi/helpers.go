package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const adapterActorHeader = "X-Adapter-Actor"

const (
	defaultPlayerDisplayName    = "player"
	defaultCommunityDisplayName = "community"
)

func stableID(kind string, value string) string {
	sum := sha256.Sum256([]byte(kind + ":" + value))
	return kind + "_" + hex.EncodeToString(sum[:])[:12]
}

func optionalProfile(r *http.Request, config ServerConfig) *MeResponse {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		return nil
	}
	if _, ok := config.Auth.Verify(token); !ok {
		return nil
	}
	profile, _, _, _, err := resolveActorProfile(r, config)
	if err != nil {
		return nil
	}
	AddRequestLogFields(r.Context(), slog.String("actor_type", "player"), slog.String("player_id", profile.Player.ID))
	return &profile
}

func resolveActorProfile(r *http.Request, config ServerConfig) (MeResponse, int, string, string, error) {
	platform, serverID, userID, err := parseAdapterActor(r.Header.Get(adapterActorHeader))
	if err != nil {
		return MeResponse{}, http.StatusBadRequest, "missing_adapter_actor", "missing adapter actor", err
	}

	resolved, err := config.Store.ResolveExternalActor(
		r.Context(),
		platform,
		serverID,
		userID,
		defaultPlayerDisplayName,
		defaultCommunityDisplayName,
	)
	if err != nil {
		return MeResponse{}, http.StatusInternalServerError, "resolve_external_actor_failed", "resolve external actor", err
	}

	player, community, err := loadResolvedProfile(r, config, resolved.PlayerID, resolved.CommunityID)
	if err != nil {
		return MeResponse{}, http.StatusInternalServerError, "load_resolved_profile_failed", "load resolved profile", err
	}

	return MeResponse{Player: player, Community: community}, 0, "", "", nil
}

func loadResolvedProfile(r *http.Request, config ServerConfig, playerID string, communityID string) (Player, Community, error) {
	player, ok, err := config.Store.FindPlayer(r.Context(), playerID)
	if err != nil {
		return Player{}, Community{}, err
	}
	if !ok {
		return Player{}, Community{}, fmt.Errorf("resolved player %q not found", playerID)
	}

	community, ok, err := config.Store.FindCommunity(r.Context(), communityID)
	if err != nil {
		return Player{}, Community{}, err
	}
	if !ok {
		return Player{}, Community{}, fmt.Errorf("resolved community %q not found", communityID)
	}

	return Player{ID: player.ID, DisplayName: player.DisplayName}, Community{ID: community.ID, DisplayName: community.DisplayName}, nil
}

func parseAdapterActor(header string) (platform string, serverID string, userID string, err error) {
	parts := strings.Split(strings.TrimSpace(header), ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("adapter actor must be platform:server:user")
	}
	platform = strings.TrimSpace(parts[0])
	serverID = strings.TrimSpace(parts[1])
	userID = strings.TrimSpace(parts[2])
	if platform == "" || serverID == "" || userID == "" {
		return "", "", "", fmt.Errorf("adapter actor parts must be non-empty")
	}
	return platform, serverID, userID, nil
}
