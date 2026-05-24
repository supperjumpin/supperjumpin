import assert from "node:assert/strict";
import test from "node:test";

import { getMe } from "./index.js";

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
