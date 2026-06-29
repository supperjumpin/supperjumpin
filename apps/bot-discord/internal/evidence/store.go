package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type StoreConfig struct {
	Dir     string
	BaseURL string
}

type Store struct {
	dir     string
	baseURL string
}

func NewStore(cfg StoreConfig) (*Store, error) {
	if cfg.Dir == "" {
		return nil, fmt.Errorf("evidence: StoreConfig.Dir is required")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("evidence: StoreConfig.BaseURL is required")
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("evidence: mkdir %q: %w", cfg.Dir, err)
	}
	return &Store{dir: cfg.Dir, baseURL: strings.TrimRight(cfg.BaseURL, "/")}, nil
}

func (s *Store) Save(ctx context.Context, sourceURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("evidence: build save request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("evidence: fetch source: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("evidence: source returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("evidence: read source body: %w", err)
	}

	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	ext := filepath.Ext(sourceURL)
	name := hash + ext
	path := filepath.Join(s.dir, name)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", fmt.Errorf("evidence: write file: %w", err)
	}
	return s.baseURL + "/evidence/" + name, nil
}

func (s *Store) FileServer() http.Handler {
	fs := http.FileServer(http.Dir(s.dir))
	return http.StripPrefix("/evidence/", fs)
}
