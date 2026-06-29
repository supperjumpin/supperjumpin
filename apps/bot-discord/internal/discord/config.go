package discord

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken        string
	AdapterToken    string
	APIBaseURL      string
	DataDir         string
	EvidenceAddr    string
	EvidenceBaseURL string
	AppID           string
	GuildID         string
}

func ConfigFromEnv() (Config, error) {
	cfg := Config{
		BotToken:        os.Getenv("SUPPERJUMPIN_BOT_TOKEN"),
		AdapterToken:    os.Getenv("SUPPERJUMPIN_ADAPTER_TOKEN"),
		APIBaseURL:      os.Getenv("SUPPERJUMPIN_API_BASE_URL"),
		DataDir:         envOrDefault("SUPPERJUMPIN_BOT_DATA_DIR", ".bot-data"),
		EvidenceAddr:    envOrDefault("SUPPERJUMPIN_BOT_EVIDENCE_ADDR", ":9999"),
		EvidenceBaseURL: envOrDefault("SUPPERJUMPIN_BOT_EVIDENCE_BASE_URL", "http://localhost:9999"),
		AppID:           os.Getenv("SUPPERJUMPIN_BOT_APP_ID"),
		GuildID:         os.Getenv("SUPPERJUMPIN_BOT_GUILD_ID"),
	}
	if cfg.BotToken == "" {
		return Config{}, fmt.Errorf("discord: SUPPERJUMPIN_BOT_TOKEN is required")
	}
	if cfg.AdapterToken == "" {
		return Config{}, fmt.Errorf("discord: SUPPERJUMPIN_ADAPTER_TOKEN is required")
	}
	if cfg.APIBaseURL == "" {
		return Config{}, fmt.Errorf("discord: SUPPERJUMPIN_API_BASE_URL is required")
	}
	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
