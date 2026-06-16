package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
)

func TestCreatePerformedJump_TDD_Stress(t *testing.T) {
	server := newTestServer(t)

	aliceToken := "alice-token"

	t.Run("Successful Creation", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Best jump ever",
			"mediaObjectKey": "media/123",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			ID string `json:"id"`
		}
		decodeResponse(t, rec, &resp)

		if resp.ID == "" {
			t.Fatal("expected created jump to have an ID")
		}
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		body := map[string]string{
			"source":      "Taco Bell",
			"destination": "Olive Garden",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for missing fields, got %d", rec.Code)
		}
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Best jump ever",
			"mediaObjectKey": "media/123",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", "invalid-token", body)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 for invalid token, got %d", rec.Code)
		}
	})

	t.Run("Duplicate evidence object key does not create a second visible Jump", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Retry-safe jump",
			"mediaObjectKey": "media/dup-123",
		}

		afterFirst := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if afterFirst.Code != http.StatusCreated {
			t.Fatalf("expected first create to succeed, got %d: %s", afterFirst.Code, afterFirst.Body.String())
		}
		var first struct {
			ID string `json:"id"`
		}
		decodeResponse(t, afterFirst, &first)

		second := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if second.Code != http.StatusCreated {
			t.Fatalf("expected retry to return the existing Jump, got %d: %s", second.Code, second.Body.String())
		}
		var retry struct {
			ID string `json:"id"`
		}
		decodeResponse(t, second, &retry)
		if retry.ID != first.ID {
			t.Fatalf("expected retry to return existing Jump %q, got %q", first.ID, retry.ID)
		}

		feedRec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
		if feedRec.Code != http.StatusOK {
			t.Fatalf("expected feed request to succeed, got %d: %s", feedRec.Code, feedRec.Body.String())
		}

		var feed struct {
			Jumps []struct {
				Source         string `json:"source"`
				Destination    string `json:"destination"`
				Food           string `json:"food"`
				Caption        string `json:"caption"`
				MediaObjectKey string `json:"mediaObjectKey"`
			} `json:"jumps"`
		}
		decodeResponse(t, feedRec, &feed)

		count := 0
		for _, jump := range feed.Jumps {
			if jump.Source == body["source"] && jump.Destination == body["destination"] && jump.Food == body["food"] && jump.Caption == body["caption"] {
				count++
				if jump.MediaObjectKey != body["mediaObjectKey"] {
					t.Fatalf("expected mediaObjectKey %q, got %q", body["mediaObjectKey"], jump.MediaObjectKey)
				}
			}
		}
		if count != 1 {
			t.Fatalf("expected one visible Jump after retry, got %d", count)
		}
	})

	t.Run("Concurrent duplicate evidence object key returns one Jump", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Concurrent retry-safe jump",
			"mediaObjectKey": "media/concurrent-123",
		}

		type result struct {
			code int
			id   string
			err  string
		}
		results := make(chan result, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		for range 2 {
			go func() {
				defer wg.Done()
				rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
				var resp struct {
					ID string `json:"id"`
				}
				if rec.Code == http.StatusCreated {
					if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&resp); err != nil {
						results <- result{code: rec.Code, err: err.Error()}
						return
					}
				}
				results <- result{code: rec.Code, id: resp.ID}
			}()
		}
		wg.Wait()
		close(results)

		var firstID string
		for res := range results {
			if res.err != "" {
				t.Fatalf("expected valid JSON response, got %s", res.err)
			}
			if res.code != http.StatusCreated {
				t.Fatalf("expected concurrent create to return 201, got %d", res.code)
			}
			if firstID == "" {
				firstID = res.id
				continue
			}
			if res.id != firstID {
				t.Fatalf("expected concurrent retries to return the same Jump ID, got %q and %q", firstID, res.id)
			}
		}
	})

	t.Run("Successful Off-Season Jump", func(t *testing.T) {
		body := map[string]string{
			"source":         "Taco Bell",
			"destination":    "Olive Garden",
			"food":           "Crunchwrap",
			"caption":        "Best jump ever",
			"mediaObjectKey": "media/456",
		}
		rec := doJSON(server, http.MethodPost, "/v1/jumps", aliceToken, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 for jump, got %d", rec.Code)
		}

		var resp struct {
			ID string `json:"id"`
		}
		decodeResponse(t, rec, &resp)

		if resp.ID == "" {
			t.Fatal("expected created jump to have an ID")
		}
	})
}
