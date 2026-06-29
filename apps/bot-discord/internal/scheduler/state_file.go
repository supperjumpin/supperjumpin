package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type JSONStateFile struct {
	Path string
}

func (f *JSONStateFile) Read() (map[string]time.Time, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]time.Time{}, nil
		}
		return nil, fmt.Errorf("scheduler: read state file: %w", err)
	}
	if len(data) == 0 {
		return map[string]time.Time{}, nil
	}
	var stored map[string]time.Time
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("scheduler: parse state file: %w", err)
	}
	return stored, nil
}

func (f *JSONStateFile) Write(snapshot map[string]time.Time) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return fmt.Errorf("scheduler: mkdir state dir: %w", err)
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("scheduler: marshal state: %w", err)
	}
	tmp := f.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("scheduler: write state file: %w", err)
	}
	if err := os.Rename(tmp, f.Path); err != nil {
		return fmt.Errorf("scheduler: rename state file: %w", err)
	}
	return nil
}

func (f *JSONStateFile) Remove(roundID string) error {
	stored, err := f.Read()
	if err != nil {
		return err
	}
	if _, ok := stored[roundID]; !ok {
		return nil
	}
	delete(stored, roundID)
	return f.Write(stored)
}
