package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStackProcessArtifacts(t *testing.T) {
	artifacts := stackProcessArtifacts("api")
	if artifacts.PIDPath != repoPath(".mage", "api.pid") {
		t.Fatalf("PIDPath = %q, want %q", artifacts.PIDPath, repoPath(".mage", "api.pid"))
	}
	if artifacts.LogPath != repoPath(".mage", "api.log") {
		t.Fatalf("LogPath = %q, want %q", artifacts.LogPath, repoPath(".mage", "api.log"))
	}
}

func TestStopProcessRemovesPIDForAlreadyFinishedProcess(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bot.pid")
	if err := os.WriteFile(p, []byte("123"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := stopProcessArtifacts(processArtifacts{PIDPath: p}, func(pid int, signal os.Signal) error {
		if pid != 123 {
			t.Fatalf("pid = %d, want 123", pid)
		}
		return os.ErrProcessDone
	})
	if err != nil {
		t.Fatalf("stopProcessArtifacts() error = %v", err)
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("pid file should be removed, stat err = %v", err)
	}
}
