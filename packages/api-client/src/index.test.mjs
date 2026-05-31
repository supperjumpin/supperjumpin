     1|import assert from "node:assert/strict";
     2|import test from "node:test";
     3|
     4|import {
     5|  acceptInvite,
     6|  authorizeEvidenceUpload,
     7|  createGroup,
     8|  createIdea,
     9|  createInvite,
    10|  createPlannedStunt,
    11|  getGroupHome,
    12|  getMe,
    13|  listGroups,
    14|  startSeason,
    15|  submitEvidence,
    16|  submitJudgment,
    17|} from "./index.js";
    18|
    19|test("getMe calls the backend with the Supabase bearer token", async () => {
    20|  const seen = {};
    21|  const me = await getMe({
    22|    baseUrl: "http://api.example.test",
    23|    accessToken: "supabase-access-token",
    24|    fetchImpl: async (url, init) => {
    25|      seen.url = url;
    26|      seen.authorization = init.headers.Authorization;
    27|      return Response.json({
    28|        account: { id: "account_123", email: "player@example.com" },
    29|        player: { id: "player_123", displayName: "player" },
    30|      });
    31|    },
    32|  });
    33|
    34|  assert.equal(seen.url, "http://api.example.test/v1/me");
    35|  assert.equal(seen.authorization, "Bearer supabase-access-token");
    36|  assert.equal(me.account.id, "account_123");
    37|  assert.equal(me.player.id, "player_123");
    38|});
    39|
    40|test("createGroup calls the backend with the Group name and bearer token", async () => {
    41|  const seen = {};
    42|  const home = await createGroup({
    43|    baseUrl: "http://api.example.test",
    44|    accessToken: "supabase-access-token",
    45|    name: "Breakfast Crew",
    46|    fetchImpl: async (url, init) => {
    47|      seen.url = url;
    48|      seen.method = init.method;
    49|      seen.authorization = init.headers.Authorization;
    50|      seen.body = JSON.parse(init.body);
    51|      return Response.json(
    52|        groupHomeResponse({ id: "group_123", name: "Breakfast Crew" }),
    53|        { status: 201 },
    54|      );
    55|    },
    56|  });
    57|
    58|  assert.equal(seen.url, "http://api.example.test/v1/groups");
    59|  assert.equal(seen.method, "POST");
    60|  assert.equal(seen.authorization, "Bearer supabase-access-token");
    61|  assert.deepEqual(seen.body, { name: "Breakfast Crew" });
    62|  assert.equal(home.group.name, "Breakfast Crew");
    63|  assert.equal(home.membership.role, "Group Admin");
    64|});
    65|
    66|test("listGroups returns the signed-in Player's Group Memberships", async () => {
    67|  const seen = {};
    68|  const groups = await listGroups({
    69|    baseUrl: "http://api.example.test",
    70|    accessToken: "supabase-access-token",
    71|    fetchImpl: async (url, init) => {
    72|      seen.url = url;
    73|      seen.authorization = init.headers.Authorization;
    74|      return Response.json({
    75|        memberships: [
    76|          {
    77|            group: { id: "group_123", name: "Breakfast Crew" },
    78|            membership: { groupId: "group_123", playerId: "player_123", role: "Group Admin" },
    79|          },
    80|        ],
    81|      });
    82|    },
    83|  });
    84|
    85|  assert.equal(seen.url, "http://api.example.test/v1/groups");
    86|  assert.equal(seen.authorization, "Bearer supabase-access-token");
    87|  assert.equal(groups.memberships[0].group.name, "Breakfast Crew");
    88|});
    89|
    90|test("getGroupHome fetches recent Performed Stunts for a selected Group", async () => {
    91|  const seen = {};
    92|  const home = await getGroupHome({
    93|    baseUrl: "http://api.example.test",
    94|    accessToken: "supabase-access-token",
    95|    groupId: "group_123",
    96|    fetchImpl: async (url, init) => {
    97|      seen.url = url;
    98|      seen.authorization = init.headers.Authorization;
    99|      return Response.json(
   100|        groupHomeResponse(
   101|          { id: "group_123", name: "Breakfast Crew" },
   102|          null,
   103|          [
   104|            {
   105|              stunt: stuntResponse({ status: "Performed Stunt", seasonId: "season_123", offSeason: false }),
   106|              performer: { id: "player_123", displayName: "alice" },
   107|              evidence: {
   108|                id: "evidence_123",
   109|                stuntId: "stunt_123",
   110|                caption: "Crunchwrap successfully smuggled into the parking lot.",
   111|                mediaObjectKey: "evidence_object_123",
   112|                createdAt: "2026-06-01T00:00:00Z",
   113|              },
   114|            },
   115|          ],
   116|        ),
   117|      );
   118|    },
   119|  });
   120|
   121|  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/home");
   122|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   123|  assert.equal(home.activeSeason, null);
   124|  assert.equal(home.recentStunts[0].performer.displayName, "alice");
   125|  assert.equal(home.recentStunts[0].evidence.caption, "Crunchwrap successfully smuggled into the parking lot.");
   126|  assert.deepEqual(home.standings, []);
   127|});
   128|
   129|test("createInvite requests a Group Invite for the signed-in Player", async () => {
   130|  const seen = {};
   131|  const invite = await createInvite({
   132|    baseUrl: "http://api.example.test",
   133|    accessToken: "supabase-access-token",
   134|    groupId: "group_123",
   135|    fetchImpl: async (url, init) => {
   136|      seen.url = url;
   137|      seen.method = init.method;
   138|      seen.authorization = init.headers.Authorization;
   139|      return Response.json(
   140|        { id: "invite_123", groupId: "group_123", token: "invite-token", createdBy: "player_123", expiresAt: "2026-06-01T00:00:00Z" },
   141|        { status: 201 },
   142|      );
   143|    },
   144|  });
   145|
   146|  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/invites");
   147|  assert.equal(seen.method, "POST");
   148|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   149|  assert.equal(invite.token, "invite-token");
   150|});
   151|
   152|test("acceptInvite returns the invited Group home", async () => {
   153|  const seen = {};
   154|  const home = await acceptInvite({
   155|    baseUrl: "http://api.example.test",
   156|    accessToken: "supabase-access-token",
   157|    token: "invite-token",
   158|    fetchImpl: async (url, init) => {
   159|      seen.url = url;
   160|      seen.method = init.method;
   161|      seen.authorization = init.headers.Authorization;
   162|      return Response.json(groupHomeResponse({ id: "group_123", name: "Breakfast Crew" }));
   163|    },
   164|  });
   165|
   166|  assert.equal(seen.url, "http://api.example.test/v1/invites/invite-token/accept");
   167|  assert.equal(seen.method, "POST");
   168|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   169|  assert.equal(home.group.name, "Breakfast Crew");
   170|});
   171|
   172|test("startSeason creates an Active Season for a Group through the backend", async () => {
   173|  const seen = {};
   174|  const home = await startSeason({
   175|    baseUrl: "http://api.example.test",
   176|    accessToken: "supabase-access-token",
   177|    groupId: "group_123",
   178|    fetchImpl: async (url, init) => {
   179|      seen.url = url;
   180|      seen.method = init.method;
   181|      seen.authorization = init.headers.Authorization;
   182|      return Response.json(
   183|        groupHomeResponse(
   184|          { id: "group_123", name: "Breakfast Crew" },
   185|          { id: "season_123", groupId: "group_123", commissionerPlayerId: "player_123", status: "Active" },
   186|        ),
   187|        { status: 201 },
   188|      );
   189|    },
   190|  });
   191|
   192|  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/seasons");
   193|  assert.equal(seen.method, "POST");
   194|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   195|  assert.equal(home.activeSeason.status, "Active");
   196|  assert.equal(home.activeSeason.commissionerPlayerId, "player_123");
   197|});
   198|
   199|test("createIdea posts Source Destination and Food for a Group", async () => {
   200|  const seen = {};
   201|  const idea = await createIdea({
   202|    baseUrl: "http://api.example.test",
   203|    accessToken: "supabase-access-token",
   204|    groupId: "group_123",
   205|    source: "Taco Bell",
   206|    destination: "Olive Garden parking lot",
   207|    food: "Crunchwrap",
   208|    fetchImpl: async (url, init) => {
   209|      seen.url = url;
   210|      seen.method = init.method;
   211|      seen.authorization = init.headers.Authorization;
   212|      seen.body = JSON.parse(init.body);
   213|      return Response.json(
   214|        stuntResponse({ status: "Idea", seasonId: null, offSeason: true }),
   215|        { status: 201 },
   216|      );
   217|    },
   218|  });
   219|
   220|  assert.equal(seen.url, "http://api.example.test/v1/groups/group_123/ideas");
   221|  assert.equal(seen.method, "POST");
   222|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   223|  assert.deepEqual(seen.body, {
   224|    source: "Taco Bell",
   225|    destination: "Olive Garden parking lot",
   226|    food: "Crunchwrap",
   227|  });
   228|  assert.equal(idea.status, "Idea");
   229|});
   230|
   231|test("createPlannedStunt can request default Season-linked or explicit Off-Season behavior", async () => {
   232|  const calls = [];
   233|  const planned = await createPlannedStunt({
   234|    baseUrl: "http://api.example.test",
   235|    accessToken: "supabase-access-token",
   236|    ideaId: "stunt_123",
   237|    fetchImpl: async (url, init) => {
   238|      calls.push({ url, method: init.method, authorization: init.headers.Authorization, body: init.body });
   239|      return Response.json(
   240|        stuntResponse({ status: "Planned Stunt", seasonId: "season_123", offSeason: false }),
   241|        { status: 201 },
   242|      );
   243|    },
   244|  });
   245|  const offSeason = await createPlannedStunt({
   246|    baseUrl: "http://api.example.test",
   247|    accessToken: "supabase-access-token",
   248|    ideaId: "stunt_456",
   249|    offSeason: true,
   250|    fetchImpl: async (url, init) => {
   251|      calls.push({ url, method: init.method, authorization: init.headers.Authorization, body: init.body });
   252|      return Response.json(
   253|        stuntResponse({ id: "stunt_456", status: "Planned Stunt", seasonId: null, offSeason: true }),
   254|        { status: 201 },
   255|      );
   256|    },
   257|  });
   258|
   259|  assert.equal(calls[0].url, "http://api.example.test/v1/ideas/stunt_123/planned-stunt");
   260|  assert.equal(calls[0].method, "POST");
   261|  assert.equal(calls[0].authorization, "Bearer supabase-access-token");
   262|  assert.equal(calls[0].body, undefined);
   263|  assert.equal(planned.seasonId, "season_123");
   264|  assert.equal(planned.offSeason, false);
   265|  assert.equal(calls[1].url, "http://api.example.test/v1/ideas/stunt_456/planned-stunt");
   266|  assert.deepEqual(JSON.parse(calls[1].body), { offSeason: true });
   267|  assert.equal(offSeason.seasonId, null);
   268|  assert.equal(offSeason.offSeason, true);
   269|});
   270|
   271|test("authorizeEvidenceUpload requests a direct upload target for a Planned Stunt", async () => {
   272|  const seen = {};
   273|  const authorization = await authorizeEvidenceUpload({
   274|    baseUrl: "http://api.example.test",
   275|    accessToken: "supabase-access-token",
   276|    stuntId: "stunt_123",
   277|    contentType: "image/jpeg",
   278|    fetchImpl: async (url, init) => {
   279|      seen.url = url;
   280|      seen.method = init.method;
   281|      seen.authorization = init.headers.Authorization;
   282|      seen.body = JSON.parse(init.body);
   283|      return Response.json(
   284|        {
   285|          id: "evidence_upload_123",
   286|          stuntId: "stunt_123",
   287|          uploadUrl: "https://storage.supperjumpin.test/uploads/evidence_object_123",
   288|          uploadMethod: "PUT",
   289|          uploadHeaders: { "Content-Type": "image/jpeg" },
   290|          mediaObjectKey: "evidence_object_123",
   291|          expiresAt: "2026-06-01T00:15:00Z",
   292|        },
   293|        { status: 201 },
   294|      );
   295|    },
   296|  });
   297|
   298|  assert.equal(seen.url, "http://api.example.test/v1/stunts/stunt_123/evidence-upload-authorizations");
   299|  assert.equal(seen.method, "POST");
   300|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   301|  assert.deepEqual(seen.body, { contentType: "image/jpeg" });
   302|  assert.equal(authorization.uploadMethod, "PUT");
   303|  assert.equal(authorization.mediaObjectKey, "evidence_object_123");
   304|});
   305|
   306|test("submitEvidence finalizes backend-owned Evidence for a Planned Stunt", async () => {
   307|  const seen = {};
   308|  const submission = await submitEvidence({
   309|    baseUrl: "http://api.example.test",
   310|    accessToken: "supabase-access-token",
   311|    stuntId: "stunt_123",
   312|    uploadAuthorizationId: "evidence_upload_123",
   313|    caption: "Crunchwrap successfully smuggled into the parking lot.",
   314|    fetchImpl: async (url, init) => {
   315|      seen.url = url;
   316|      seen.method = init.method;
   317|      seen.authorization = init.headers.Authorization;
   318|      seen.body = JSON.parse(init.body);
   319|      return Response.json(
   320|        {
   321|          stunt: stuntResponse({ status: "Performed Stunt" }),
   322|          evidence: {
   323|            id: "evidence_123",
   324|            stuntId: "stunt_123",
   325|            caption: "Crunchwrap successfully smuggled into the parking lot.",
   326|            mediaObjectKey: "evidence_object_123",
   327|            createdAt: "2026-06-01T00:00:00Z",
   328|          },
   329|        },
   330|        { status: 201 },
   331|      );
   332|    },
   333|  });
   334|
   335|  assert.equal(seen.url, "http://api.example.test/v1/stunts/stunt_123/evidence");
   336|  assert.equal(seen.method, "POST");
   337|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   338|  assert.deepEqual(seen.body, {
   339|    uploadAuthorizationId: "evidence_upload_123",
   340|    caption: "Crunchwrap successfully smuggled into the parking lot.",
   341|  });
   342|  assert.equal(submission.stunt.status, "Performed Stunt");
   343|  assert.equal(submission.evidence.mediaObjectKey, "evidence_object_123");
   344|});
   345|
   346|test("submitJudgment posts the four Judgment scores for a Performed Stunt", async () => {
   347|  const seen = {};
   348|  const judgment = await submitJudgment({
   349|    baseUrl: "http://api.example.test",
   350|    accessToken: "supabase-access-token",
   351|    stuntId: "stunt_123",
   352|    commitment: 4,
   353|    transgression: 5,
   354|    creativity: 3,
   355|    documentation: 2,
   356|    fetchImpl: async (url, init) => {
   357|      seen.url = url;
   358|      seen.method = init.method;
   359|      seen.authorization = init.headers.Authorization;
   360|      seen.body = JSON.parse(init.body);
   361|      return Response.json(
   362|        {
   363|          id: "judgment_123",
   364|          stuntId: "stunt_123",
   365|          playerId: "player_456",
   366|          commitment: 4,
   367|          transgression: 5,
   368|          creativity: 3,
   369|          documentation: 2,
   370|        },
   371|        { status: 201 },
   372|      );
   373|    },
   374|  });
   375|
   376|  assert.equal(seen.url, "http://api.example.test/v1/stunts/stunt_123/judgment");
   377|  assert.equal(seen.method, "POST");
   378|  assert.equal(seen.authorization, "Bearer supabase-access-token");
   379|  assert.deepEqual(seen.body, {
   380|    commitment: 4,
   381|    transgression: 5,
   382|    creativity: 3,
   383|    documentation: 2,
   384|  });
   385|  assert.equal(judgment.playerId, "player_456");
   386|  assert.equal(judgment.transgression, 5);
   387|});
   388|
   389|function groupHomeResponse(group, activeSeason = null, recentStunts = []) {
   390|  return {
   391|    group,
   392|    membership: { groupId: group.id, playerId: "player_123", role: "Group Admin" },
   393|    activeSeason,
   394|    recentStunts,
   395|    standings: [],
   396|  };
   397|}
   398|
   399|function stuntResponse(overrides = {}) {
   400|  return {
   401|    id: "stunt_123",
   402|    groupId: "group_123",
   403|    playerId: "player_123",
   404|    seasonId: null,
   405|    status: "Idea",
   406|    source: "Taco Bell",
   407|    destination: "Olive Garden parking lot",
   408|    food: "Crunchwrap",
   409|    offSeason: true,
   410|    ...overrides,
   411|  };
   412|}
   413|