package bot

import (
	"context"
	"net/http/httptest"
	"testing"
)

type testHarness struct {
	api     *fakeAPI
	server  *httptest.Server
	client  *APIClient
	handler *RoundStartHandler
}

func newTestHarness(t *testing.T) *testHarness {
	t.Helper()
	fake := &fakeAPI{}
	server := httptest.NewServer(fake.handler())
	t.Cleanup(server.Close)

	client, err := NewAPIClient(APIClientConfig{
		BaseURL:      server.URL,
		AdapterToken: "dev-token",
		HTTPClient:   server.Client(),
	})
	if err != nil {
		t.Fatalf("new API client: %v", err)
	}

	return &testHarness{
		api:     fake,
		server:  server,
		client:  client,
		handler: NewRoundStartHandler(client, testActorResolver("discord:guild-1:user-1")),
	}
}

func (h *testHarness) handleInteraction(t *testing.T, i IncomingInteraction) (Reply, error) {
	t.Helper()
	return h.handler.Handle(context.Background(), i)
}

type capturedResponse struct {
	body      string
	ephemeral bool
}

type fakeResponder struct {
	captured    []capturedResponse
	followUps   []FollowUpMessage
}

func (f *fakeResponder) Respond(_ context.Context, reply Reply) error {
	f.captured = append(f.captured, capturedResponse{
		body:      reply.Body,
		ephemeral: reply.Ephemeral,
	})
	return nil
}

func (f *fakeResponder) PostFollowUp(_ context.Context, msg FollowUpMessage) error {
	f.followUps = append(f.followUps, msg)
	return nil
}
