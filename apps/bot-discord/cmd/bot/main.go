package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/supperjumpin/supperjumpin/apps/bot-discord/internal/discord"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := discord.ConfigFromEnv()
	if err != nil {
		logger.Error("bot: config", "err", err)
		os.Exit(1)
	}

	wired, err := discord.NewWired(cfg)
	if err != nil {
		logger.Error("bot: build", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := wired.RecoverAndStartWatchdog(ctx); err != nil {
		logger.Error("bot: recover", "err", err)
		os.Exit(1)
	}

	if err := wired.LoadStampTemplate(ctx); err != nil {
		logger.Warn("bot: load stamp template", "err", err)
	}

	go func() {
		if err := wired.StartEvidenceServer(); err != nil {
			logger.Error("bot: evidence server", "err", err)
		}
	}()
	defer wired.StopEvidenceServer()

	if err := wired.Session.Open(); err != nil {
		logger.Error("bot: open session", "err", err)
		os.Exit(1)
	}
	defer wired.Session.Close()

	if cfg.AppID != "" {
		if err := wired.RegisterCommands(cfg.AppID, cfg.GuildID); err != nil {
			logger.Warn("bot: register commands", "err", err)
		} else {
			logger.Info("bot: registered commands", "app_id", cfg.AppID, "guild_id", cfg.GuildID)
		}
	} else {
		logger.Warn("bot: SUPPERJUMPIN_BOT_APP_ID not set, skipping command registration")
	}

	logger.Info("Supperjumpin Discord bot connected",
		"api_base_url", cfg.APIBaseURL,
		"evidence_addr", cfg.EvidenceAddr,
	)

	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "Supperjumpin Discord bot shutting down")
}
