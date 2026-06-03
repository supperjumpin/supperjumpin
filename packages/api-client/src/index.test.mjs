import assert from "node:assert/strict";
import test from "node:test";

import {
  authorizeEvidenceUpload,
  getJumpDetail,
  getMe,
  getPublicFeed,
  submitEvidence,
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

test("authorizeEvidenceUpload requests a direct upload target for a Planned Jump", async () => {
  const seen = {};
  const authorization = await authorizeEvidenceUpload({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    jumpId: "jump_123",
    contentType: "image/jpeg",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          id: "evidence_upload_123",
          jumpId: "jump_123",
          uploadUrl: "https://storage.supperjumpin.test/uploads/evidence_object_123",
          uploadMethod: "PUT",
          uploadHeaders: { "Content-Type": "image/jpeg" },
          mediaObjectKey: "evidence_object_123",
          expiresAt: "2026-06-01T00:15:00Z",
        },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/jumps/jump_123/evidence-upload-authorizations");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, { contentType: "image/jpeg" });
  assert.equal(authorization.uploadMethod, "PUT");
  assert.equal(authorization.mediaObjectKey, "evidence_object_123");
});

test("submitEvidence finalizes backend-owned Evidence for a Planned Jump", async () => {
  const seen = {};
  const submission = await submitEvidence({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    jumpId: "jump_123",
    uploadAuthorizationId: "evidence_upload_123",
    caption: "Crunchwrap successfully smuggled into the parking lot.",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          jump: jumpResponse({ status: "Performed Jump" }),
          evidence: {
            id: "evidence_123",
            jumpId: "jump_123",
            caption: "Crunchwrap successfully smuggled into the parking lot.",
            mediaObjectKey: "evidence_object_123",
            createdAt: "2026-06-01T00:00:00Z",
          },
        },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/jumps/jump_123/evidence");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, {
    uploadAuthorizationId: "evidence_upload_123",
    caption: "Crunchwrap successfully smuggled into the parking lot.",
  });
  assert.equal(submission.jump.status, "Performed Jump");
  assert.equal(submission.evidence.mediaObjectKey, "evidence_object_123");
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
