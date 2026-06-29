package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStore_SaveDownloadsBytesAndReturnsURL(t *testing.T) {
	payload := []byte("fake-photo-bytes")
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer source.Close()

	dir := t.TempDir()
	store, err := NewStore(StoreConfig{
		Dir:     dir,
		BaseURL: "http://localhost:9999",
	})
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	got, err := store.Save(t.Context(), source.URL+"/photo.png")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	wantExt := ".png"
	if !bytes.HasSuffix([]byte(got), []byte(wantExt)) {
		t.Errorf("URL extension: got %q, want suffix %q", got, wantExt)
	}
	if !bytes.HasPrefix([]byte(got), []byte("http://localhost:9999/evidence/")) {
		t.Errorf("URL prefix: got %q, want prefix %q", got, "http://localhost:9999/evidence/")
	}
}

func TestStore_SaveWritesToCorrectPath(t *testing.T) {
	payload := []byte("fake-photo-bytes")
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer source.Close()

	dir := t.TempDir()
	store, err := NewStore(StoreConfig{
		Dir:     dir,
		BaseURL: "http://localhost:9999",
	})
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	got, err := store.Save(t.Context(), source.URL+"/photo.png")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	wantHash := sha256.Sum256(payload)
	wantName := hex.EncodeToString(wantHash[:]) + ".png"
	wantPath := dir + "/" + wantName
	if got, want := got, "http://localhost:9999/evidence/"+wantName; got != want {
		t.Errorf("URL: got %q, want %q", got, want)
	}

	body, err := readFile(wantPath)
	if err != nil {
		t.Fatalf("readFile(%q): %v", wantPath, err)
	}
	if !bytes.Equal(body, payload) {
		t.Errorf("file contents: got %q, want %q", body, payload)
	}
}

func TestStore_FileServerServesSavedFiles(t *testing.T) {
	payload := []byte("fake-photo-bytes")
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer source.Close()

	dir := t.TempDir()
	store, err := NewStore(StoreConfig{
		Dir:     dir,
		BaseURL: "http://localhost:9999",
	})
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	url, err := store.Save(t.Context(), source.URL+"/photo.png")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	ts := httptest.NewServer(store.FileServer())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/evidence/" + filenameFromURL(url))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}

	got := make([]byte, len(payload))
	n, _ := resp.Body.Read(got)
	if !bytes.Equal(got[:n], payload) {
		t.Errorf("served body: got %q, want %q", got[:n], payload)
	}
}

func TestStore_SaveReturnsErrorOnSourceFailure(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer source.Close()

	dir := t.TempDir()
	store, err := NewStore(StoreConfig{
		Dir:     dir,
		BaseURL: "http://localhost:9999",
	})
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	_, err = store.Save(t.Context(), source.URL+"/missing.png")
	if err == nil {
		t.Fatal("Save: got nil error, want error for 404 source")
	}
}
