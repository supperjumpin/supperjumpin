package scheduler

import (
	"path/filepath"
	"testing"
	"time"
)

func TestJSONStateFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "active-reveals.json")
	sf := &JSONStateFile{Path: path}

	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	want := map[string]time.Time{
		"round-1": now.Add(1 * time.Hour),
		"round-2": now.Add(2 * time.Hour),
	}

	if err := sf.Write(want); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := sf.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Read: got %d entries, want 2", len(got))
	}
	if !got["round-1"].Equal(want["round-1"]) {
		t.Errorf("round-1: got %v, want %v", got["round-1"], want["round-1"])
	}
	if !got["round-2"].Equal(want["round-2"]) {
		t.Errorf("round-2: got %v, want %v", got["round-2"], want["round-2"])
	}
}

func TestJSONStateFile_ReadMissingReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")
	sf := &JSONStateFile{Path: path}

	got, err := sf.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Read: got %d entries, want 0", len(got))
	}
}

func TestJSONStateFile_RemoveDeletesEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "active-reveals.json")
	sf := &JSONStateFile{Path: path}

	now := time.Now()
	_ = sf.Write(map[string]time.Time{
		"round-1": now.Add(1 * time.Hour),
		"round-2": now.Add(2 * time.Hour),
	})

	if err := sf.Remove("round-1"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	got, err := sf.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if _, ok := got["round-1"]; ok {
		t.Errorf("round-1: still present, want removed")
	}
	if _, ok := got["round-2"]; !ok {
		t.Errorf("round-2: missing, want present")
	}
}
