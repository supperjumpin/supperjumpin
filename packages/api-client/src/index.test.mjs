import assert from "node:assert/strict";
import test from "node:test";

import { acceptInvite, createGroup, createInvite, getGroupHome, getMe, listGroups } from "./index.js";

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

test("getGroupHome fetches backend Season Stunts and Standings placeholders for a selected Group", async () => {
  const seen = {};
  const home = await getGroupHome({
    baseUrl: "http://api.example.test",
    accessToken: "supabase-access-token",
    groupId: "group_123",
    fetchImpl: async (url, init) => {
      seen.url = url;
      seen.authorization = init.headers.Authorization;
      return Response.json(groupHomeResponse({ id: "group_123", name: "Breakfast Crew" }));
    },
  });

  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/home");
  assert.equal(seen.authorization, "Bearer supabase-access-token");
  assert.equal(home.activeSeason, null);
  assert.deepEqual(home.recentStunts, []);
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

function groupHomeResponse(group) {
  return {
    group,
    membership: { groupId: group.id, playerId: "player_123", role: "Group Admin" },
    activeSeason: null,
    recentStunts: [],
    standings: [],
  };
}
