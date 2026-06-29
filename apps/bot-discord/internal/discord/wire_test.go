package discord

import (
	"testing"
)

func TestConfigFromEnv_RequiresAllFields(t *testing.T) {
	clearEnv(t)

	_, err := ConfigFromEnv()
	if err == nil {
		t.Fatal("ConfigFromEnv: got nil error, want error when SUPPERJUMPIN_BOT_TOKEN, SUPPERJUMPIN_ADAPTER_TOKEN, and SUPPERJUMPIN_API_BASE_URL are all unset")
	}
}

func TestConfigFromEnv_ReadsAllFields(t *testing.T) {
	clearEnv(t)
	t.Setenv("SUPPERJUMPIN_BOT_TOKEN", "bot-token-xyz")
	t.Setenv("SUPPERJUMPIN_ADAPTER_TOKEN", "adapter-token-abc")
	t.Setenv("SUPPERJUMPIN_API_BASE_URL", "http://localhost:8080")

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}
	if cfg.BotToken != "bot-token-xyz" {
		t.Errorf("BotToken: got %q, want %q", cfg.BotToken, "bot-token-xyz")
	}
	if cfg.AdapterToken != "adapter-token-abc" {
		t.Errorf("AdapterToken: got %q, want %q", cfg.AdapterToken, "adapter-token-abc")
	}
	if cfg.APIBaseURL != "http://localhost:8080" {
		t.Errorf("APIBaseURL: got %q, want %q", cfg.APIBaseURL, "http://localhost:8080")
	}
}

func TestWired_BuildsBotAndSessionFromConfig(t *testing.T) {
	clearEnv(t)
	t.Setenv("SUPPERJUMPIN_BOT_TOKEN", "bot-token-xyz")
	t.Setenv("SUPPERJUMPIN_ADAPTER_TOKEN", "adapter-token-abc")
	t.Setenv("SUPPERJUMPIN_API_BASE_URL", "http://localhost:8080")
	t.Setenv("SUPPERJUMPIN_BOT_DATA_DIR", t.TempDir())

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}

	wired, err := NewWired(cfg)
	if err != nil {
		t.Fatalf("NewWired: %v", err)
	}
	if wired.Bot == nil {
		t.Error("Wired.Bot: nil, want non-nil")
	}
	if wired.Dispatcher == nil {
		t.Error("Wired.Dispatcher: nil, want non-nil")
	}
	if wired.Session == nil {
		t.Error("Wired.Session: nil, want non-nil")
	}
	if wired.Scheduler == nil {
		t.Error("Wired.Scheduler: nil, want non-nil")
	}
	if got, want := wired.Session.Token, "Bot bot-token-xyz"; got != want {
		t.Errorf("Session.Token: got %q, want %q", got, want)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SUPPERJUMPIN_BOT_TOKEN", "")
	t.Setenv("SUPPERJUMPIN_ADAPTER_TOKEN", "")
	t.Setenv("SUPPERJUMPIN_API_BASE_URL", "")
	t.Setenv("SUPPERJUMPIN_BOT_DATA_DIR", "")
}
