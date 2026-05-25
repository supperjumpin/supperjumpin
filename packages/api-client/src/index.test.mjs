import assert from "node:assert/strict";
import test from "node:test";

import {
  acceptInvite,
  authorizeEvidenceUpload,
  createGroup,
  createIdea,
  createInvite,
  createPlannedStunt,
  getGroupHome,
  getMe,
  listGroups,
  startSeason,
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

test("createGroup calls the backend with the Group name and bearer token", async () => {
  const seen = {};
  const home = await createGroup({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    name: "Breakfast Crew",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        groupHomeResponse({ id: "group_123", name: "Breakfast Crew" }),
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, { name: "Breakfast Crew" });
  assert.equal(home.group.name, "Breakfast Crew");
  assert.equal(home.membership.role, "Group Admin");
});

test("listGroups returns the signed-in Player's Group Memberships", async () => {
  const seen = {};
  const groups = await listGroups({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.authorization = init.headers.Authorization;
      return Response.json({
        memberships: [
          {
            group: { id: "group_123", name: "Breakfast Crew" },
            membership: { groupId: "group_123", playerId: "player_123", role: "Group Admin" },
          },
        ],
      });
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(groups.memberships[0].group.name, "Breakfast Crew");
});

test("getGroupHome fetches recent Performed Stunts for a selected Group", async () => {
  const seen = {};
  const home = await getGroupHome({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    groupId: "group_123",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.authorization = init.headers.Authorization;
      return Response.json(
        groupHomeResponse(
          { id: "group_123", name: "Breakfast Crew" },
          null,
          [
            {
              stunt: stuntResponse({ status: "Performed Stunt", seasonId: "season_123", offSeason: false }),
              performer: { id: "player_123", displayName: "alice" },
              evidence: {
                id: "evidence_123",
                stuntId: "stunt_123",
                caption: "Crunchwrap successfully smuggled into the parking lot.",
                mediaObjectKey: "evidence_object_123",
                createdAt: "2026-06-01T00:00:00Z",
              },
            },
          ],
        ),
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/home");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(home.activeSeason, null);
  assert.equal(home.recentStunts[0].performer.displayName, "alice");
  assert.equal(home.recentStunts[0].evidence.caption, "Crunchwrap successfully smuggled into the parking lot.");
  assert.deepEqual(home.standings, []);
});

test("createInvite requests a Group Invite for the signed-in Player", async () => {
  const seen = {};
  const invite = await createInvite({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    groupId: "group_123",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      return Response.json(
        { id: "invite_123", groupId: "group_123", token: "invite-token", createdBy: "player_123", expiresAt: "2026-06-01T00:00:00Z" },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/invites");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(invite.token, "invite-token");
});

test("acceptInvite returns the invited Group home", async () => {
  const seen = {};
  const home = await acceptInvite({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    token: "invite-token",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      return Response.json(groupHomeResponse({ id: "group_123", name: "Breakfast Crew" }));
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/invites/invite-token/accept");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(home.group.name, "Breakfast Crew");
});

test("startSeason creates an Active Season for a Group through the backend", async () => {
  const seen = {};
  const home = await startSeason({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    groupId: "group_123",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      return Response.json(
        groupHomeResponse(
          { id: "group_123", name: "Breakfast Crew" },
          { id: "season_123", groupId: "group_123", commissionerPlayerId: "player_123", status: "Active" },
        ),
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/seasons");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(home.activeSeason.status, "Active");
  assert.equal(home.activeSeason.commissionerPlayerId, "player_123");
});

test("createIdea posts Source Destination and Food for a Group", async () => {
  const seen = {};
  const idea = await createIdea({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    groupId: "group_123",
    source: "Taco Bell",
    destination: "Olive Garden parking lot",
    food: "Crunchwrap",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        stuntResponse({ status: "Idea", seasonId: null, offSeason: true }),
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/ideas");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, {
    source: "Taco Bell",
    destination: "Olive Garden parking lot",
    food: "Crunchwrap",
  });
  assert.equal(idea.status, "Idea");
});

test("createPlannedStunt can request default Season-linked or explicit Off-Season behavior", async () => {
  const calls = [];
  const planned = await createPlannedStunt({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    ideaId: "stunt_123",
    fetchImpl: async (url, init) => {
      calls.push({ url, method: init.method, authorization: init.headers.Authorization, body: init.body });
      return Response.json(
        stuntResponse({ status: "Planned Stunt", seasonId: "season_123", offSeason: false }),
        { status: 201 },
      );
    },
  });
  const offSeason = await createPlannedStunt({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    ideaId: "stunt_456",
    offSeason: true,
    fetchImpl: async (url, init) => {
      calls.push({ url, method: init.method, authorization: init.headers.Authorization, body: init.body });
      return Response.json(
        stuntResponse({ id: "stunt_456", status: "Planned Stunt", seasonId: null, offSeason: true }),
        { status: 201 },
      );
    },
  });

  assert.equal(calls[0].url, "http://api.example.test/v1/ideas/stunt_123/planned-stunt");
  assert.equal(calls[0].method, "POST");
  assert.equal(calls[0].authorization, "Bearer supabase-access-token");
  assert.equal(calls[0].body, undefined);
  assert.equal(planned.seasonId, "season_123");
  assert.equal(planned.offSeason, false);
  assert.equal(calls[1].url, "http://api.example.test/v1/ideas/stunt_456/planned-stunt");
  assert.deepEqual(JSON.parse(calls[1].body), { offSeason: true });
  assert.equal(offSeason.seasonId, null);
  assert.equal(offSeason.offSeason, true);
});

test("authorizeEvidenceUpload requests a direct upload target for a Planned Stunt", async () => {
  const seen = {};
  const authorization = await authorizeEvidenceUpload({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    stuntId: "stunt_123",
    contentType: "image/jpeg",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          id: "evidence_upload_123",
          stuntId: "stunt_123",
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

  assert.equal(seen.url, "http://api.example.test/v1/stunts/stunt_123/evidence-upload-authorizations");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, { contentType: "image/jpeg" });
  assert.equal(authorization.uploadMethod, "PUT");
  assert.equal(authorization.mediaObjectKey, "evidence_object_123");
});

test("submitEvidence finalizes backend-owned Evidence for a Planned Stunt", async () => {
  const seen = {};
  const submission = await submitEvidence({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    stuntId: "stunt_123",
    uploadAuthorizationId: "evidence_upload_123",
    caption: "Crunchwrap successfully smuggled into the parking lot.",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          stunt: stuntResponse({ status: "Performed Stunt" }),
          evidence: {
            id: "evidence_123",
            stuntId: "stunt_123",
            caption: "Crunchwrap successfully smuggled into the parking lot.",
            mediaObjectKey: "evidence_object_123",
            createdAt: "2026-06-01T00:00:00Z",
          },
        },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/stunts/stunt_123/evidence");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, {
    uploadAuthorizationId: "evidence_upload_123",
    caption: "Crunchwrap successfully smuggled into the parking lot.",
  });
  assert.equal(submission.stunt.status, "Performed Stunt");
  assert.equal(submission.evidence.mediaObjectKey, "evidence_object_123");
});

test("submitJudgment posts the four Judgment scores for a Performed Stunt", async () => {
  const seen = {};
  const judgment = await submitJudgment({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    stuntId: "stunt_123",
    difficulty: 4,
    transgression: 5,
    creativity: 3,
    documentation: 2,
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.method = init.method;
      seen.authorization = init.headers.Authorization;
      seen.body = JSON.parse(init.body);
      return Response.json(
        {
          id: "judgment_123",
          stuntId: "stunt_123",
          playerId: "player_456",
          difficulty: 4,
          transgression: 5,
          creativity: 3,
          documentation: 2,
        },
        { status: 201 },
      );
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/stunts/stunt_123/judgment");
  assert.equal(seen.method, "POST");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.deepEqual(seen.body, {
    difficulty: 4,
    transgression: 5,
    creativity: 3,
    documentation: 2,
  });
  assert.equal(judgment.playerId, "player_456");
  assert.equal(judgment.transgression, 5);
});

function groupHomeResponse(group, activeSeason = null, recentStunts = []) {
  return {
    group,
    membership: { groupId: group.id, playerId: "player_123", role: "Group Admin" },
    activeSeason,
    recentStunts,
    standings: [],
  };
}

function stuntResponse(overrides = {}) {
  return {
    id: "stunt_123",
    groupId: "group_123",
    playerId: "player_123",
    seasonId: null,
    status: "Idea",
    source: "Taco Bell",
    destination: "Olive Garden parking lot",
    food: "Crunchwrap",
    offSeason: true,
    ...overrides,
  };
}
