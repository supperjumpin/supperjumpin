package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/benbjohnson/clock"
	discordgo "github.com/bwmarrin/discordgo"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/bot"
	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/evidence"
	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/scheduler"
)

type Wired struct {
	APIClient   *bot.APIClient
	Bot         *bot.Bot
	Dispatcher  *Dispatcher
	Session     *discordgo.Session
	Scheduler   *scheduler.Scheduler
	Evidence    *evidence.Store
	EvidenceSrv *http.Server
	Registry    *bot.RoundRegistry
	Renderer    *Renderer
}

func NewWired(cfg Config) (*Wired, error) {
	apiClient, err := bot.NewAPIClient(bot.APIClientConfig{
		BaseURL:      cfg.APIBaseURL,
		AdapterToken: cfg.AdapterToken,
	})
	if err != nil {
		return nil, fmt.Errorf("discord: build API client: %w", err)
	}

	resolver := bot.ActorResolverFunc(func(_ context.Context, guildID, userID string) (bot.Actor, error) {
		return bot.Actor{
			GuildID: guildID,
			UserID:  userID,
			Tuple:   "discord:" + guildID + ":" + userID,
		}, nil
	})

	statePath := filepath.Join(cfg.DataDir, "active-reveals.json")
	stateFile := &scheduler.JSONStateFile{Path: statePath}

	evidenceDir := filepath.Join(cfg.DataDir, "evidence")
	evidenceStore, err := evidence.NewStore(evidence.StoreConfig{
		Dir:     evidenceDir,
		BaseURL: cfg.EvidenceBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("discord: build evidence store: %w", err)
	}

	registry := bot.NewRoundRegistry()
	poster := NewSessionChannelPoster(nil) // wired later, after session build
	renderer := NewRenderer(poster)

	revealActor := &bot.RevealActor{
		Client:   apiClient,
		Resolver: resolver,
		Registry: registry,
		Poster:   renderer,
	}

	revealSched := scheduler.New(scheduler.Config{
		Clock:  clock.New(),
		OnFire: revealActor.Fire,
		State:  stateFile,
	})

	roundStart := bot.NewRoundStartHandler(apiClient, resolver)
	roundStart.SetRevealScheduler(revealSched)
	roundStart.SetRegistry(registry)
	roundStatus := bot.NewRoundStatusHandler(apiClient, resolver)
	jumpSubmit := bot.NewJumpSubmitHandler(apiClient, evidenceStore, resolver)
	stampApply := bot.NewStampApplyHandler(apiClient, resolver)
	commit := bot.NewCommitHandler(apiClient, resolver)
	comment := bot.NewCommentHandler(apiClient, resolver)
	recap := bot.NewRecapHandler(apiClient, resolver)

	b := bot.NewBot(bot.BotConfig{
		RoundStart:  roundStart.AsHandlerFunc(),
		RoundStatus: roundStatus.AsHandlerFunc(),
		JumpSubmit:  jumpSubmit.AsHandlerFunc(),
		StampApply:  stampApply.AsHandlerFunc(),
		Commit:      commit.AsHandlerFunc(),
		Comment:     comment.AsHandlerFunc(),
		Recap:       recap.AsHandlerFunc(),
	})

	dispatcher := NewDispatcher(b, WithLogger(slog.Default()))

	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("discord: build session: %w", err)
	}

	poster = NewSessionChannelPoster(session)

	session.AddHandler(func(s *discordgo.Session, event *discordgo.InteractionCreate) {
		dispatcher.HandleInteraction(s, event)
	})

	evidenceSrv := &http.Server{
		Addr:    cfg.EvidenceAddr,
		Handler: evidenceHTTPHandler(evidenceStore),
	}

	return &Wired{
		APIClient:   apiClient,
		Bot:         b,
		Dispatcher:  dispatcher,
		Session:     session,
		Scheduler:   revealSched,
		Evidence:    evidenceStore,
		EvidenceSrv: evidenceSrv,
		Registry:    registry,
		Renderer:    renderer,
	}, nil
}

func evidenceHTTPHandler(store *evidence.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("/evidence/", store.FileServer())
	return mux
}

func (w *Wired) LoadStampTemplate(ctx context.Context) error {
	if w.APIClient == nil || w.Renderer == nil {
		return fmt.Errorf("discord: stamp template dependencies not wired")
	}

	resp, err := w.APIClient.ListStampCatalog(ctx)
	if err != nil {
		return fmt.Errorf("discord: list stamp catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord: list stamp catalog status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Stamps []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Glyph string `json:"glyph"`
		} `json:"stamps"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("discord: decode stamp catalog: %w", err)
	}

	stamps := make([]StampTemplate, 0, len(parsed.Stamps))
	for _, s := range parsed.Stamps {
		stamps = append(stamps, StampTemplate{ID: s.ID, Label: s.Label, Glyph: s.Glyph})
	}
	w.Renderer.SetStampTemplate(stamps)
	return nil
}

func (w *Wired) RegisterCommands(appID, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
		RoundStartCommand(),
		JumpSubmitCommand(),
		CommentCommand(),
		RecapCommand(),
	}
	return RegisterCommands(w.Session, appID, guildID, commands...)
}

func (w *Wired) RecoverAndStartWatchdog(ctx context.Context) error {
	if err := w.Scheduler.Recover(ctx); err != nil {
		return fmt.Errorf("discord: recover reveals: %w", err)
	}
	w.Scheduler.StartWatchdog(1 * time.Minute)
	return nil
}

func (w *Wired) StartEvidenceServer() error {
	if err := w.EvidenceSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("discord: evidence server: %w", err)
	}
	return nil
}

func (w *Wired) StopEvidenceServer() error {
	return w.EvidenceSrv.Close()
}
