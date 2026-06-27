package httpapi_test

import (
	"context"
	"testing"
)

func TestResolveExternalActorIsIdempotent(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	ctx := context.Background()

	result1, err := store.ResolveExternalActor(ctx, "discord", "guild-123", "user-abc", "coolkoala", "Supper Club")
	if err != nil {
		t.Fatalf("first call: unexpected error %v", err)
	}
	if !result1.Created {
		t.Fatal("first call: expected Created=true")
	}
	if result1.PlayerID == "" {
		t.Fatal("first call: expected non-empty PlayerID")
	}
	if result1.CommunityID == "" {
		t.Fatal("first call: expected non-empty CommunityID")
	}

	result2, err := store.ResolveExternalActor(ctx, "discord", "guild-123", "user-abc", "coolkoala", "Supper Club")
	if err != nil {
		t.Fatalf("second call: unexpected error %v", err)
	}
	if result2.Created {
		t.Fatal("second call: expected Created=false")
	}
	if result2.PlayerID != result1.PlayerID {
		t.Fatalf("second call: expected same PlayerID (%q != %q)", result1.PlayerID, result2.PlayerID)
	}
	if result2.CommunityID != result1.CommunityID {
		t.Fatalf("second call: expected same CommunityID (%q != %q)", result1.CommunityID, result2.CommunityID)
	}
}

func TestResolveExternalActorTwoActorsSameServer(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	ctx := context.Background()

	alice, err := store.ResolveExternalActor(ctx, "discord", "guild-123", "user-alice", "alice", "Supper Club")
	if err != nil {
		t.Fatalf("alice: unexpected error %v", err)
	}
	if !alice.Created {
		t.Fatal("alice: expected Created=true")
	}

	bob, err := store.ResolveExternalActor(ctx, "discord", "guild-123", "user-bob", "bob", "Supper Club")
	if err != nil {
		t.Fatalf("bob: unexpected error %v", err)
	}
	if !bob.Created {
		t.Fatal("bob: expected Created=true")
	}

	if bob.CommunityID != alice.CommunityID {
		t.Fatalf("expected same community for same server (%q != %q)", alice.CommunityID, bob.CommunityID)
	}
	if bob.PlayerID == alice.PlayerID {
		t.Fatalf("expected different player IDs for different users (both %q)", alice.PlayerID)
	}
}

func TestResolveExternalActorDifferentServerDifferentCommunity(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	ctx := context.Background()

	serverA, err := store.ResolveExternalActor(ctx, "discord", "guild-a", "user-1", "player1", "Club A")
	if err != nil {
		t.Fatalf("server A: unexpected error %v", err)
	}

	serverB, err := store.ResolveExternalActor(ctx, "discord", "guild-b", "user-1", "player1", "Club B")
	if err != nil {
		t.Fatalf("server B: unexpected error %v", err)
	}

	if serverB.CommunityID == serverA.CommunityID {
		t.Fatal("expected different community IDs for different servers")
	}
	if serverB.PlayerID == serverA.PlayerID {
		t.Fatal("expected different player IDs for different servers (same user)")
	}
}
