import assert from "node:assert/strict";
import test from "node:test";

import {
  createJump,
  getJumpDetail,
  getMe,
  getPublicFeed,
  updateDisplayName,
  submitJudgment,
} from "./index.js";

test("getMe calls the backend with the Supabase bearer token", async () => {
  const seen = {};
  const me = await getMe({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.authorization = init.headers.Authorization;
      return Response.json({
        account: { id: "account_123", email: "player@example.com" },
        player: { id: "player_123", displayName: "player" },
      });
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/me");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(me.account.id, "account_123");
  assert.equal(me.player.id, "player_123");
});

test("submitJudgment posts the four Judgment scores for a Performed Jump", async () => {
  const seen = {};
  const judgment = await submitJudgment({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    jumpId: "jump_123",
    commitment: 4,
    transgression: 5,
    creativity: 3,
    presentation: 2,
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          id: "judgment_123",
          jumpId: "jump_123",
          playerId: "player_456",
          commitment: 4,
          transgression: 5,
          creativity: 3,
          presentation: 2,
        },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/jumps/jump_123/judgment");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, {
    commitment: 4,
    transgression: 5,
    creativity: 3,
    presentation: 2,
  });
  assert.equal(judgment.playerId, "player_456");
  assert.equal(judgment.transgression, 5);
});

test("createJump posts the required Jump fields with bearer auth", async () => {
  const seen = {};
  const jump = await createJump({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    source: "Taco Bell",
    destination: "Olive Garden",
    food: "Crunchwrap",
    caption: "Best jump ever",
    mediaObjectKey: "evidence/jump_123.jpg",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          id: "jump_123",
          playerId: "player_456",
          status: "Performed Jump",
          source: "Taco Bell",
          destination: "Olive Garden",
          food: "Crunchwrap",
          finalScore: null,
          openFinalScore: null,
          gracePeriodExpiresAt: "2026-06-16T12:10:00Z",
          createdAt: "2026-06-16T12:00:00Z",
        },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/jumps");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, {
    source: "Taco Bell",
    destination: "Olive Garden",
    food: "Crunchwrap",
    caption: "Best jump ever",
    mediaObjectKey: "evidence/jump_123.jpg",
  });
  assert.equal(jump.id, "jump_123");
  assert.equal(jump.status, "Performed Jump");
});

test("updateDisplayName patches the player's display name with bearer auth", async () => {
  const seen = {};
  const response = await updateDisplayName({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    displayName: "new-handle",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json({
        player: { id: "player_123", displayName: "new-handle" },
      });
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/me/display-name");
  assert.equal(seen.method, "PATCH");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, { displayName: "new-handle" });
  assert.equal(response.player.displayName, "new-handle");
});

test("getPublicFeed fetches the public feed without requiring auth", async () => {
  const seen = {};
  const response = await getPublicFeed({
    baseUrl: "http://api.example.test",
    cursor: "cursor_123",
    limit: 10,
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.authorization = init.headers.Authorization;
      return Response.json({
        jumps: [
          {
            id: "jump_123",
            performerId: "player_123",
            performerName: "alice",
            source: "Taco Bell",
            destination: "Olive Garden parking lot",
            food: "Crunchwrap",
            caption: "Crunchwrap successfully smuggled into the parking lot.",
            mediaObjectKey: "evidence_object_123",
            status: "Performed Jump",
            gracePeriodExpiresAt: "2026-06-01T00:10:00Z",
            runningAverage: 3.5,
            judgmentCount: 4,
            createdAt: "2026-06-01T00:00:00Z",
          },
        ],
        nextCursor: "cursor_456",
      });
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/feed?cursor=cursor_123&limit=10");
  assert.equal(seen.authorization, undefined);
  assert.equal(response.jumps[0].performerName, "alice");
  assert.equal(response.nextCursor, "cursor_456");
});

test("getJumpDetail includes bearer auth when a viewer token is present", async () => {
  const seen = {};
  const detail = await getJumpDetail({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    jumpId: "jump_123",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.authorization = init.headers.Authorization;
      return Response.json({
        id: "jump_123",
        performerId: "player_123",
        performerName: "alice",
        source: "Taco Bell",
        destination: "Olive Garden parking lot",
        food: "Crunchwrap",
        caption: "Crunchwrap successfully smuggled into the parking lot.",
        mediaObjectKey: "evidence_object_123",
        status: "Performed Jump",
        gracePeriodExpiresAt: "2026-06-01T00:10:00Z",
        runningAverage: 3.5,
        judgmentCount: 4,
        createdAt: "2026-06-01T00:00:00Z",
        viewerContext: { canJudge: false, reason: "already-judged", hasJudged: true },
      });
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/jumps/jump_123");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(detail.id, "jump_123");
  assert.equal(detail.viewerContext.reason, "already-judged");
});

test("public read helpers surface backend message fields in thrown errors", async () => {
  await assert.rejects(
    () =>
      getPublicFeed({
        baseUrl: "http://api.example.test",
        fetchImpl: async () =>
          Response.json(
            { error: "internal_error", message: "Could not load jumps. Please try again." },
            { status: 500 },
          ),
      }),
    (err) => {
      assert.equal(err.message, "Could not load jumps. Please try again.");
      return true;
    },
  );

  await assert.rejects(
    () =>
      getJumpDetail({
        baseUrl: "http://api.example.test",
        jumpId: "jump_123",
        fetchImpl: async () =>
          Response.json(
            { error: "not_found", message: "Jump not found. It may have been removed." },
            { status: 404 },
          ),
      }),
    (err) => {
      assert.equal(err.message, "Jump not found. It may have been removed.");
      return true;
    },
  );
});

function jumpResponse(overrides = {}) {
  return {
    id: "jump_123",
    playerId: "player_123",
    status: "Performed Jump",
    source: "Taco Bell",
    destination: "Olive Garden parking lot",
    food: "Crunchwrap",
    ...overrides,
  };
}
