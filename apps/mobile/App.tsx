     1|import { GoTrueClient } from "@supabase/auth-js";
     2|import { useEffect, useState } from "react";
     3|import { Button, SafeAreaView, ScrollView, StyleSheet, Text, TextInput, View, PanResponder } from "react-native";
     4|
     5|import {
     6|  acceptInvite,
     7|  createGroup,
     8|  createIdea,
     9|  createInvite,
    10|  createPlannedJump,
    11|  getGroupHome,
    12|  getMe,
    13|  listGroups,
    14|  startSeason,
    15|  submitJudgment,
    16|} from "@supperjumpin/api-client";
    17|import type {
    18|  GroupHomeResponse,
    19|  GroupMembershipSummary,
    20|  Judgment,
    21|  MeResponse,
    22|  PerformedJumpView,
    23|  Jump,
    24|} from "@supperjumpin/api-client";
    25|
    26|const supabaseUrl = process.env.EXPO_PUBLIC_SUPABASE_URL ?? "https://example.supabase.co";
    27|const supabaseAnonKey = process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY ?? "";
    28|const supabase = {
    29|  auth: new GoTrueClient({
    30|    url: `${supabaseUrl.replace(/\/$/, "")}/auth/v1`,
    31|    headers: { apikey: supabaseAnonKey },
    32|  }),
    33|};
    34|
    35|const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
    36|
    37|export default function App() {
    38|  const [email, setEmail] = useState("");
    39|  const [groupName, setGroupName] = useState("");
    40|  const [inviteToken, setInviteToken] = useState("");
    41|  const [source, setSource] = useState("");
    42|  const [destination, setDestination] = useState("");
    43|  const [food, setFood] = useState("");
    44|  const [status, setStatus] = useState("Enter an email to request a Supabase magic link.");
    45|  const [accessToken, setAccessToken] = useState<string | null>(null);
    46|  const [profile, setProfile] = useState<MeResponse | null>(null);
    47|  const [memberships, setMemberships] = useState<GroupMembershipSummary[]>([]);
    48|  const [groupHome, setGroupHome] = useState<GroupHomeResponse | null>(null);
    49|  const [selectedJump, setSelectedJump] = useState<PerformedJumpView | null>(null);
    50|  const [idea, setIdea] = useState<Jump | null>(null);
    51|  const [plannedJump, setPlannedJump] = useState<Jump | null>(null);
    52|  const [gestureScores, setGestureScores] = useState<{ commitment: number; transgression: number; creativity: number; presentation: number } | null>(null);
    53|  const [pendingJudgment, setPendingJudgment] = useState<Judgment | null>(null);
    54|  const [isJudging, setIsJudging] = useState(false);
    55|
    56|  useEffect(() => {
    57|    supabase.auth.getSession().then(async ({ data }) => {
    58|      if (!data.session) {
    59|        return;
    60|      }
    61|
    62|      const token = data.session.access_token;
    63|      setAccessToken(token);
    64|      const me = await getMe({ baseUrl: apiBaseUrl, accessToken: token });
    65|      const groups = await listGroups({ baseUrl: apiBaseUrl, accessToken: token });
    66|      setProfile(me);
    67|      setMemberships(groups.memberships);
    68|      if (groups.memberships[0]) {
    69|        await selectGroup(token, groups.memberships[0].group.id);
    70|      }
    71|      setStatus("Signed in and connected to Supperjumpin.");
    72|    });
    73|  }, []);
    74|
    75|  async function sendMagicLink() {
    76|    const { error } = await supabase.auth.signInWithOtp({ email });
    77|    setStatus(error ? error.message : "Check your email for the Supperjumpin magic link.");
    78|  }
    79|
    80|  async function createNewGroup() {
    81|    if (!accessToken) {
    82|      setStatus("Sign in before creating a Group.");
    83|      return;
    84|    }
    85|
    86|    const home = await createGroup({ baseUrl: apiBaseUrl, accessToken, name: groupName });
    87|    const groups = await listGroups({ baseUrl: apiBaseUrl, accessToken });
    88|    setGroupHome(home);
    89|    setSelectedJump(home.recentJumps[0] ?? null);
    90|    setMemberships(groups.memberships);
    91|    setGroupName("");
    92|    setStatus(`Created ${home.group.name}. You are its Group Admin.`);
    93|  }
    94|
    95|  async function createGroupInvite() {
    96|    if (!accessToken || !groupHome) {
    97|      setStatus("Select a Group before creating an Invite.");
    98|      return;
    99|    }
   100|
   101|    try {
   102|      const invite = await createInvite({ baseUrl: apiBaseUrl, accessToken, groupId: groupHome.group.id });
   103|      setInviteToken(invite.token);
   104|      setStatus(`Created Invite for ${groupHome.group.name}.`);
   105|    } catch (error) {
   106|      setStatus(error instanceof Error ? error.message : "Could not create Invite.");
   107|    }
   108|  }
   109|
   110|  async function acceptGroupInvite() {
   111|    if (!accessToken) {
   112|      setStatus("Sign in before accepting an Invite.");
   113|      return;
   114|    }
   115|
   116|    try {
   117|      const home = await acceptInvite({ baseUrl: apiBaseUrl, accessToken, token: inviteToken.trim() });
   118|      const groups = await listGroups({ baseUrl: apiBaseUrl, accessToken });
   119|      setGroupHome(home);
   120|      setSelectedJump(home.recentJumps[0] ?? null);
   121|      setMemberships(groups.memberships);
   122|      setInviteToken("");
   123|      setStatus(`Joined ${home.group.name}.`);
   124|    } catch (error) {
   125|      setStatus(error instanceof Error ? error.message : "Could not accept Invite.");
   126|    }
   127|  }
   128|
   129|  async function selectGroup(token: string, groupId: string) {
   130|    const home = await getGroupHome({ baseUrl: apiBaseUrl, accessToken: token, groupId });
   131|    setGroupHome(home);
   132|    setSelectedJump(home.recentJumps[0] ?? null);
   133|    setIdea(null);
   134|    setPlannedJump(null);
   135|  }
   136|
   137|  async function startNewSeason() {
   138|    if (!accessToken || !groupHome) {
   139|      setStatus("Select a Group before starting a Season.");
   140|      return;
   141|    }
   142|
   143|    const home = await startSeason({ baseUrl: apiBaseUrl, accessToken, groupId: groupHome.group.id });
   144|    setGroupHome(home);
   145|    setSelectedJump(home.recentJumps[0] ?? null);
   146|    setStatus(`Started an Active Season for ${home.group.name}. You are the Season Commissioner.`);
   147|  }
   148|
   149|  async function captureIdea() {
   150|    if (!accessToken || !groupHome) {
   151|      setStatus("Select a Group before creating an Idea.");
   152|      return;
   153|    }
   154|
   155|    const captured = await createIdea({
   156|      baseUrl: apiBaseUrl,
   157|      accessToken,
   158|      groupId: groupHome.group.id,
   159|      source,
   160|      destination,
   161|      food,
   162|    });
   163|    setIdea(captured);
   164|    setPlannedJump(null);
   165|    setStatus(`Captured Idea: ${captured.food} from ${captured.source} at ${captured.destination}.`);
   166|  }
   167|
   168|  async function planJump(offSeason = false) {
   169|    if (!accessToken || !idea) {
   170|      setStatus("Create an Idea before creating a Planned Jump.");
   171|      return;
   172|    }
   173|
   174|    const planned = await createPlannedJump({ baseUrl: apiBaseUrl, accessToken, ideaId: idea.id, offSeason });
   175|    setPlannedJump(planned);
   176|    setIdea(null);
   177|    setSource("");
   178|    setDestination("");
   179|    setFood("");
   180|    setStatus(
   181|      planned.offSeason
   182|        ? `Created an Off-Season Jump for ${planned.food}.`
   183|        : `Created a Season-linked Planned Jump for ${planned.food}.`,
   184|    );
   185|  }
   186|
   187|  async function startJudging(jump: PerformedJumpView) {
   188|    if (!accessToken) {
   189|      setStatus("Sign in to judge jumps.");
   190|      return;
   191|    }
   192|    if (jump.performer.id === profile?.player.id) {
   193|      setStatus("You cannot judge your own jump.");
   194|      return;
   195|    }
   196|    setIsJudging(true);
   197|    setGestureScores({ commitment: 5, transgression: 5, creativity: 5, presentation: 5 });
   198|    setSelectedJump(jump);
   199|    setStatus("Use gestures or buttons to set scores. Swipe up/down on each factor.");
   200|  }
   201|
   202|  function updateGestureScore(factor: "commitment" | "transgression" | "creativity" | "presentation", delta: number) {
   203|    if (!gestureScores) return;
   204|    const newValue = Math.max(0, Math.min(10, gestureScores[factor] + delta));
   205|    setGestureScores({ ...gestureScores, [factor]: newValue });
   206|  }
   207|
   208|  function clearGestureScores() {
   209|    setGestureScores({ commitment: 5, transgression: 5, creativity: 5, presentation: 5 });
   210|    setStatus("Gesture scores cleared. Start fresh.");
   211|  }
   212|
   213|  async function submitGestureJudgment() {
   214|    if (!accessToken || !selectedJump || !gestureScores) {
   215|      setStatus("Cannot submit judgment without scores.");
   216|      return;
   217|    }
   218|
   219|    try {
   220|      const judgment = await submitJudgment({
   221|        baseUrl: apiBaseUrl,
   222|        accessToken,
   223|        jumpId: selectedJump.jump.id,
   224|        ...gestureScores,
   225|      });
   226|      setPendingJudgment(judgment);
   227|      setIsJudging(false);
   228|      setGestureScores(null);
   229|      setStatus(`Judgment submitted for ${selectedJump.jump.food}.`);
   230|    } catch (error) {
   231|      setStatus(error instanceof Error ? error.message : "Could not submit judgment.");
   232|    }
   233|  }
   234|
   235|  function cancelJudging() {
   236|    setIsJudging(false);
   237|    setGestureScores(null);
   238|    setStatus("Judging cancelled.");
   239|  }
   240|
   241|  function createScorePanResponder(factor: "commitment" | "transgression" | "creativity" | "presentation") {
   242|    let lastY = 0;
   243|    return PanResponder.create({
   244|      onStartShouldSetPanResponder: () => true,
   245|      onMoveShouldSetPanResponder: (_, gestureState) => Math.abs(gestureState.dy) > 10,
   246|      onPanResponderGrant: (_, gestureState) => {
   247|        lastY = gestureState.y0;
   248|      },
   249|      onPanResponderMove: (_, gestureState) => {
   250|        const deltaY = lastY - gestureState.moveY;
   251|        if (Math.abs(deltaY) > 50) {
   252|          updateGestureScore(factor, deltaY > 0 ? 1 : -1);
   253|          lastY = gestureState.moveY;
   254|        }
   255|      },
   256|      onPanResponderRelease: () => {},
   257|    });
   258|  }
   259|
   260|  const commitmentPan = createScorePanResponder("commitment");
   261|  const transgressionPan = createScorePanResponder("transgression");
   262|  const creativityPan = createScorePanResponder("creativity");
   263|  const presentationPan = createScorePanResponder("presentation");
   264|
   265|  return (
   266|    <SafeAreaView style={styles.screen}>
   267|      <ScrollView contentContainerStyle={styles.content}>
   268|        <View style={styles.card}>
   269|        <Text style={styles.title}>Supperjumpin</Text>
   270|        <Text style={styles.body}>{status}</Text>
   271|        <TextInput
   272|          autoCapitalize="none"
   273|          keyboardType="email-address"
   274|          onChangeText={setEmail}
   275|          placeholder="player@example.com"
   276|          style={styles.input}
   277|          value={email}
   278|        />
   279|        <Button onPress={sendMagicLink} title="Send magic link" />
   280|        {profile ? (
   281|          <Text style={styles.body}>
   282|            Account {profile.account.id} is attached to Player {profile.player.id}.
   283|          </Text>
   284|        ) : null}
   285|        {profile ? (
   286|          <View style={styles.section}>
   287|            <Text style={styles.sectionTitle}>Create Group</Text>
   288|            <TextInput
   289|              onChangeText={setGroupName}
   290|              placeholder="Breakfast Crew"
   291|              style={styles.input}
   292|              value={groupName}
   293|            />
   294|            <Button onPress={createNewGroup} title="Create Group" />
   295|          </View>
   296|        ) : null}
   297|        {memberships.length > 0 ? (
   298|          <View style={styles.section}>
   299|            <Text style={styles.sectionTitle}>Your Groups</Text>
   300|            {memberships.map((entry) => (
   301|              <Button
   302|                key={entry.group.id}
   303|                onPress={() => accessToken && selectGroup(accessToken, entry.group.id)}
   304|                title={`${entry.group.name} (${entry.membership.role})`}
   305|              />
   306|            ))}
   307|          </View>
   308|        ) : null}
   309|        {profile ? (
   310|          <View style={styles.section}>
   311|            <Text style={styles.sectionTitle}>Invites</Text>
   312|            <Button onPress={createGroupInvite} title="Create Invite for selected Group" />
   313|            <TextInput
   314|              autoCapitalize="none"
   315|              onChangeText={setInviteToken}
   316|              placeholder="Invite token"
   317|              style={styles.input}
   318|              value={inviteToken}
   319|            />
   320|            <Button onPress={acceptGroupInvite} title="Accept Invite" />
   321|          </View>
   322|        ) : null}
   323|        {groupHome ? (
   324|          <View style={styles.groupHome}>
   325|            <Text style={styles.groupName}>{groupHome.group.name}</Text>
   326|            <Text style={styles.body}>Your Group Membership: {groupHome.membership.role}</Text>
   327|            <Text style={styles.body}>
   328|              Active Season: {groupHome.activeSeason ? groupHome.activeSeason.status : "None"}
   329|            </Text>
   330|            {groupHome.activeSeason ? (
   331|              <Text style={styles.body}>
   332|                Season Commissioner: {groupHome.activeSeason.commissionerPlayerId}
   333|              </Text>
   334|            ) : (
   335|              <Button onPress={startNewSeason} title="Start Season" />
   336|            )}
   337|            <Text style={styles.body}>Recent Jumps: {groupHome.recentJumps.length}</Text>
   338|            <Text style={styles.body}>Standings: {groupHome.standings.length}</Text>
   339|            {groupHome.recentJumps.length > 0 ? (
   340|              <View style={styles.section}>
   341|                <Text style={styles.sectionTitle}>Recent Performed Jumps</Text>
   342|                {groupHome.recentJumps.map((performedJump: PerformedJumpView) => (
   343|                  <Button
   344|                    key={performedJump.jump.id}
   345|                    onPress={() => setSelectedJump(performedJump)}
   346|                    title={`${performedJump.jump.food} at ${performedJump.jump.destination}`}
   347|                  />
   348|                ))}
   349|              </View>
   350|            ) : null}
   351|            {selectedJump ? (
   352|              <View style={styles.jumpCard}>
   353|                <Text style={styles.sectionTitle}>Jump Detail</Text>
   354|                <Text style={styles.body}>Source: {selectedJump.jump.source}</Text>
   355|                <Text style={styles.body}>Destination: {selectedJump.jump.destination}</Text>
   356|                <Text style={styles.body}>Food: {selectedJump.jump.food}</Text>
   357|                <Text style={styles.body}>Performer: {selectedJump.performer.displayName}</Text>
   358|                <Text style={styles.body}>Caption: {selectedJump.evidence.caption}</Text>
   359|                <Text style={styles.body}>Media object: {selectedJump.evidence.mediaObjectKey}</Text>
   360|                {pendingJudgment && selectedJump.jump.id === pendingJudgment.jumpId ? (
   361|                  <View style={styles.judgmentSummary}>
   362|                    <Text style={styles.sectionTitle}>Your Judgment</Text>
   363|                    <Text style={styles.body}>Commitment: {pendingJudgment.commitment}</Text>
   364|                    <Text style={styles.body}>Transgression: {pendingJudgment.transgression}</Text>
   365|                    <Text style={styles.body}>Creativity: {pendingJudgment.creativity}</Text>
   366|                    <Text style={styles.body}>Presentation: {pendingJudgment.presentation}</Text>
   367|                  </View>
   368|                ) : null}
   369|                {!isJudging && selectedJump.performer.id !== profile?.player.id ? (
   370|                  <Button onPress={() => startJudging(selectedJump)} title="Judge this Jump" />
   371|                ) : null}
   372|                {selectedJump.performer.id === profile?.player.id ? (
   373|                  <Text style={styles.body}>Your own jump - cannot judge.</Text>
   374|                ) : null}
   375|              </View>
   376|            ) : null}
   377|            {isJudging && gestureScores ? (
   378|              <View style={styles.judgingCard}>
   379|                <Text style={styles.sectionTitle}>Quick-Judge</Text>
   380|                <Text style={styles.body}>Swipe up/down on each factor to adjust score (0-10)</Text>
   381|                <View style={styles.scoreRow} {...commitmentPan.panHandlers} accessibilityLabel="Commitment score adjustment. Swipe up or down to change value.">
   382|                  <Text style={styles.scoreLabel}>Commitment:</Text>
   383|                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.commitment} out of 10`}>{gestureScores.commitment}</Text>
   384|                  <Button onPress={() => updateGestureScore("commitment", -1)} title="-" accessibilityLabel="Decrease commitment score by 1" />
   385|                  <Button onPress={() => updateGestureScore("commitment", 1)} title="+" accessibilityLabel="Increase commitment score by 1" />
   386|                </View>
   387|                <View style={styles.scoreRow} {...transgressionPan.panHandlers} accessibilityLabel="Transgression score adjustment. Swipe up or down to change value.">
   388|                  <Text style={styles.scoreLabel}>Transgression:</Text>
   389|                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.transgression} out of 10`}>{gestureScores.transgression}</Text>
   390|                  <Button onPress={() => updateGestureScore("transgression", -1)} title="-" accessibilityLabel="Decrease transgression score by 1" />
   391|                  <Button onPress={() => updateGestureScore("transgression", 1)} title="+" accessibilityLabel="Increase transgression score by 1" />
   392|                </View>
   393|                <View style={styles.scoreRow} {...creativityPan.panHandlers} accessibilityLabel="Creativity score adjustment. Swipe up or down to change value.">
   394|                  <Text style={styles.scoreLabel}>Creativity:</Text>
   395|                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.creativity} out of 10`}>{gestureScores.creativity}</Text>
   396|                  <Button onPress={() => updateGestureScore("creativity", -1)} title="-" accessibilityLabel="Decrease creativity score by 1" />
   397|                  <Button onPress={() => updateGestureScore("creativity", 1)} title="+" accessibilityLabel="Increase creativity score by 1" />
   398|                </View>
   399|                <View style={styles.scoreRow} {...presentationPan.panHandlers} accessibilityLabel="Presentation score adjustment. Swipe up or down to change value.">
   400|                  <Text style={styles.scoreLabel}>Presentation:</Text>
   401|                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.presentation} out of 10`}>{gestureScores.presentation}</Text>
   402|                  <Button onPress={() => updateGestureScore("presentation", -1)} title="-" accessibilityLabel="Decrease presentation score by 1" />
   403|                  <Button onPress={() => updateGestureScore("presentation", 1)} title="+" accessibilityLabel="Increase presentation score by 1" />
   404|                </View>
   405|                <View style={styles.judgmentActions}>
   406|                  <Button onPress={clearGestureScores} title="Clear" color="#c1673a" />
   407|                  <Button onPress={cancelJudging} title="Cancel" color="#888" />
   408|                  <Button onPress={submitGestureJudgment} title="Submit Judgment" color="#2f241d" />
   409|                </View>
   410|              </View>
   411|            ) : null}
   412|            <View style={styles.section}>
   413|              <Text style={styles.sectionTitle}>Create Idea</Text>
   414|              <TextInput onChangeText={setSource} placeholder="Source" style={styles.input} value={source} />
   415|              <TextInput
   416|                onChangeText={setDestination}
   417|                placeholder="Destination"
   418|                style={styles.input}
   419|                value={destination}
   420|              />
   421|              <TextInput onChangeText={setFood} placeholder="Food" style={styles.input} value={food} />
   422|              <Button onPress={captureIdea} title="Capture Idea" />
   423|              {idea ? (
   424|                <View style={styles.jumpCard}>
   425|                  <Text style={styles.body}>Idea: {idea.food}</Text>
   426|                  <Button onPress={() => planJump(false)} title="Create Planned Jump" />
   427|                  <Button onPress={() => planJump(true)} title="Create Off-Season Jump" />
   428|                </View>
   429|              ) : null}
   430|              {plannedJump ? (
   431|                <Text style={styles.body}>
   432|                  Planned Jump: {plannedJump.source} to {plannedJump.destination} with {plannedJump.food} (
   433|                  {plannedJump.offSeason ? "Off-Season Jump" : "Season-linked"})
   434|                </Text>
   435|              ) : null}
   436|            </View>
   437|          </View>
   438|        ) : null}
   439|        </View>
   440|      </ScrollView>
   441|    </SafeAreaView>
   442|  );
   443|}
   444|
   445|const styles = StyleSheet.create({
   446|  screen: {
   447|    flex: 1,
   448|    backgroundColor: "#f7efe2",
   449|  },
   450|  content: {
   451|    flexGrow: 1,
   452|    justifyContent: "center",
   453|    padding: 24,
   454|  },
   455|  card: {
   456|    backgroundColor: "#fffaf2",
   457|    borderColor: "#2f241d",
   458|    borderRadius: 24,
   459|    borderWidth: 2,
   460|    gap: 16,
   461|    padding: 24,
   462|  },
   463|  title: {
   464|    color: "#2f241d",
   465|    fontSize: 36,
   466|    fontWeight: "900",
   467|  },
   468|  body: {
   469|    color: "#4d3b31",
   470|    fontSize: 16,
   471|  },
   472|  section: {
   473|    gap: 8,
   474|  },
   475|  sectionTitle: {
   476|    color: "#2f241d",
   477|    fontSize: 18,
   478|    fontWeight: "800",
   479|  },
   480|  groupHome: {
   481|    backgroundColor: "#f7efe2",
   482|    borderRadius: 16,
   483|    gap: 8,
   484|    padding: 16,
   485|  },
   486|  groupName: {
   487|    color: "#2f241d",
   488|    fontSize: 24,
   489|    fontWeight: "900",
   490|  },
   491|  jumpCard: {
   492|    backgroundColor: "#fffaf2",
   493|    borderColor: "#c1673a",
   494|    borderRadius: 12,
   495|    borderWidth: 1,
   496|    gap: 8,
   497|    padding: 12,
   498|  },
   499|  judgingCard: {
   500|    backgroundColor: "#e8f4e8",
   501|