package discord

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestWired_LoadStampTemplateFetchesCatalogIntoRenderer(t *testing.T) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stamp-catalog" {
			t.Fatalf("path: got %q, want %q", r.URL.Path, "/v1/stamp-catalog")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stamps":[{"id":"stamp-1","label":"Approve","glyph":"✅"},{"id":"stamp-2","label":"Chaos","glyph":"🌀"}]}`))
	}))
	defer apiServer.Close()

	clearEnv(t)
	t.Setenv("SUPPERJUMPIN_BOT_TOKEN", "bot-token-xyz")
	t.Setenv("SUPPERJUMPIN_ADAPTER_TOKEN", "adapter-token-abc")
	t.Setenv("SUPPERJUMPIN_API_BASE_URL", apiServer.URL)
	t.Setenv("SUPPERJUMPIN_BOT_DATA_DIR", t.TempDir())

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}

	wired, err := NewWired(cfg)
	if err != nil {
		t.Fatalf("NewWired: %v", err)
	}

	if err := wired.LoadStampTemplate(context.Background()); err != nil {
		t.Fatalf("LoadStampTemplate: %v", err)
	}

	if got, want := len(wired.Renderer.stampTemplate), 2; got != want {
		t.Fatalf("stamp template length: got %d, want %d", got, want)
	}
	if got, want := wired.Renderer.stampTemplate[0].ID, "stamp-1"; got != want {
		t.Errorf("stampTemplate[0].ID: got %q, want %q", got, want)
	}
	if got, want := wired.Renderer.stampTemplate[1].Label, "Chaos"; got != want {
		t.Errorf("stampTemplate[1].Label: got %q, want %q", got, want)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SUPPERJUMPIN_BOT_TOKEN", "")
	t.Setenv("SUPPERJUMPIN_ADAPTER_TOKEN", "")
	t.Setenv("SUPPERJUMPIN_API_BASE_URL", "")
	t.Setenv("SUPPERJUMPIN_BOT_DATA_DIR", "")
}
