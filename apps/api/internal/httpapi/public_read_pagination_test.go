package httpapi_test

import (
	"net/http"
	"testing"
	"time"
)

func TestPublicFeedEmptyReturnsEmptyArrayAndNullCursor(t *testing.T) {
	server := newGroupsTestServer(t)

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 0 {
		t.Fatalf("expected empty feed, got %d jumps", len(res.Jumps))
	}
	if res.NextCursor != nil {
		t.Fatalf("expected nil cursor for empty feed, got %q", *res.NextCursor)
	}
}

func TestPublicFeedCursorPaginationMultiPage(t *testing.T) {
	server, store := newPublicReadTestServer(t)

	var created []string
	for i := 0; i < 25; i++ {
		minute := i
		store.Store.SetClock(func() time.Time {
			return time.Date(2026, 6, 1, 13, minute, 0, 0, time.UTC)
		})
		performed := performJump(t, server, "alice-token", store.GroupID)
		created = append(created, performed.Jump.ID)
	}

	rec := doJSON(server, http.MethodGet, "/v1/feed?limit=10", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var page1 publicFeedBody
	decodeResponse(t, rec, &page1)
	if len(page1.Jumps) != 10 {
		t.Fatalf("expected 10 jumps on page 1, got %d", len(page1.Jumps))
	}
	if page1.NextCursor == nil {
		t.Fatal("expected nextCursor on page 1")
	}
	if page1.Jumps[0].ID != created[24] || page1.Jumps[9].ID != created[15] {
		t.Fatalf("unexpected page-1 bounds: got %q..%q, want %q..%q", page1.Jumps[0].ID, page1.Jumps[9].ID, created[24], created[15])
	}

	rec2 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page1.NextCursor+"&limit=10", "", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 2, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var page2 publicFeedBody
	decodeResponse(t, rec2, &page2)
	if len(page2.Jumps) != 10 {
		t.Fatalf("expected 10 jumps on page 2, got %d", len(page2.Jumps))
	}
	if page2.NextCursor == nil {
		t.Fatal("expected nextCursor on page 2")
	}
	if page2.Jumps[0].ID != created[14] || page2.Jumps[9].ID != created[5] {
		t.Fatalf("unexpected page-2 bounds: got %q..%q, want %q..%q", page2.Jumps[0].ID, page2.Jumps[9].ID, created[14], created[5])
	}

	rec3 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page2.NextCursor+"&limit=10", "", nil)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 3, got %d: %s", rec3.Code, rec3.Body.String())
	}
	var page3 publicFeedBody
	decodeResponse(t, rec3, &page3)
	if len(page3.Jumps) != 5 {
		t.Fatalf("expected 5 jumps on page 3, got %d", len(page3.Jumps))
	}
	if page3.NextCursor != nil {
		t.Fatal("expected nil cursor on last page")
	}
	if page3.Jumps[0].ID != created[4] || page3.Jumps[4].ID != created[0] {
		t.Fatalf("unexpected page-3 bounds: got %q..%q, want %q..%q", page3.Jumps[0].ID, page3.Jumps[4].ID, created[4], created[0])
	}

	seen := map[string]bool{}
	for _, p := range []publicFeedBody{page1, page2, page3} {
		for _, j := range p.Jumps {
			if seen[j.ID] {
				t.Fatalf("duplicate jump %q across pages", j.ID)
			}
			seen[j.ID] = true
		}
	}
	if len(seen) != 25 {
		t.Fatalf("expected 25 unique jumps, got %d", len(seen))
	}
}

func TestPublicFeedSameTimestampTiebrokenByID(t *testing.T) {
	server, store := newPublicReadTestServer(t)

	store.Store.SetClock(func() time.Time {
		return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	})
	var created []string
	for i := 0; i < 5; i++ {
		performed := performJump(t, server, "alice-token", store.GroupID)
		created = append(created, performed.Jump.ID)
	}

	rec1 := doJSON(server, http.MethodGet, "/v1/feed?limit=3", "", nil)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec1.Code, rec1.Body.String())
	}
	var page1 publicFeedBody
	decodeResponse(t, rec1, &page1)
	if len(page1.Jumps) != 3 {
		t.Fatalf("expected 3 jumps on page 1, got %d", len(page1.Jumps))
	}
	if page1.NextCursor == nil {
		t.Fatal("expected nextCursor on page 1")
	}

	rec1b := doJSON(server, http.MethodGet, "/v1/feed?limit=3", "", nil)
	if rec1b.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec1b.Code, rec1b.Body.String())
	}
	var page1b publicFeedBody
	decodeResponse(t, rec1b, &page1b)
	for i := 0; i < 3; i++ {
		if page1.Jumps[i].ID != page1b.Jumps[i].ID {
			t.Fatalf("deterministic order violated at index %d: %q vs %q", i, page1.Jumps[i].ID, page1b.Jumps[i].ID)
		}
	}

	rec2 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page1.NextCursor+"&limit=3", "", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 2, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var page2 publicFeedBody
	decodeResponse(t, rec2, &page2)
	if len(page2.Jumps) != 2 {
		t.Fatalf("expected 2 jumps on page 2, got %d", len(page2.Jumps))
	}
	if page2.NextCursor != nil {
		t.Fatal("expected nil cursor on last page")
	}

	seen := map[string]bool{}
	for _, j := range page1.Jumps {
		seen[j.ID] = true
	}
	for _, j := range page2.Jumps {
		if seen[j.ID] {
			t.Fatalf("duplicate jump %q across pages", j.ID)
		}
		seen[j.ID] = true
	}
	if len(seen) != 5 {
		t.Fatalf("expected 5 unique jumps, got %d", len(seen))
	}
}
