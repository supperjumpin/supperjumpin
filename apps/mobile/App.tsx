import { GoTrueClient } from "@supabase/auth-js";
import { useEffect, useState } from "react";
import { Button, SafeAreaView, ScrollView, StyleSheet, Text, TextInput, View, PanResponder } from "react-native";

import {
  acceptInvite,
  createGroup,
  createIdea,
  createInvite,
  createPlannedJump,
  getGroupHome,
  getMe,
  listGroups,
  startSeason,
  submitJudgment,
} from "@supperjumpin/api-client";
import type {
  GroupHomeResponse,
  GroupMembershipSummary,
  Judgment,
  MeResponse,
  PerformedJumpView,
  Jump,
} from "@supperjumpin/api-client";

const supabaseUrl = process.env.EXPO_PUBLIC_SUPABASE_URL ?? "https://example.supabase.co";
const supabaseAnonKey = process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY ?? "";
const supabase = {
  auth: new GoTrueClient({
    url: `${supabaseUrl.replace(/\/$/, "")}/auth/v1`,
    headers: { apikey: supabaseAnonKey },
  }),
};

const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export default function App() {
  const [email, setEmail] = useState("");
  const [groupName, setGroupName] = useState("");
  const [inviteToken, setInviteToken] = useState("");
  const [source, setSource] = useState("");
  const [destination, setDestination] = useState("");
  const [food, setFood] = useState("");
  const [status, setStatus] = useState("Enter an email to request a Supabase magic link.");
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [profile, setProfile] = useState<MeResponse | null>(null);
  const [memberships, setMemberships] = useState<GroupMembershipSummary[]>([]);
  const [groupHome, setGroupHome] = useState<GroupHomeResponse | null>(null);
  const [selectedJump, setSelectedJump] = useState<PerformedJumpView | null>(null);
  const [idea, setIdea] = useState<Jump | null>(null);
  const [plannedJump, setPlannedJump] = useState<Jump | null>(null);
  const [gestureScores, setGestureScores] = useState<{ difficulty: number; transgression: number; creativity: number; presentation: number } | null>(null);
  const [pendingJudgment, setPendingJudgment] = useState<Judgment | null>(null);
  const [isJudging, setIsJudging] = useState(false);

  useEffect(() => {
    supabase.auth.getSession().then(async ({ data }) => {
      if (!data.session) {
        return;
      }

      const token = data.session.access_token;
      setAccessToken(token);
      const me = await getMe({ baseUrl: apiBaseUrl, accessToken: token });
      const groups = await listGroups({ baseUrl: apiBaseUrl, accessToken: token });
      setProfile(me);
      setMemberships(groups.memberships);
      if (groups.memberships[0]) {
        await selectGroup(token, groups.memberships[0].group.id);
      }
      setStatus("Signed in and connected to Supperjumpin.");
    });
  }, []);

  async function sendMagicLink() {
    const { error } = await supabase.auth.signInWithOtp({ email });
    setStatus(error ? error.message : "Check your email for the Supperjumpin magic link.");
  }

  async function createNewGroup() {
    if (!accessToken) {
      setStatus("Sign in before creating a Group.");
      return;
    }

    const home = await createGroup({ baseUrl: apiBaseUrl, accessToken, name: groupName });
    const groups = await listGroups({ baseUrl: apiBaseUrl, accessToken });
    setGroupHome(home);
    setSelectedJump(home.recentJumps[0] ?? null);
    setMemberships(groups.memberships);
    setGroupName("");
    setStatus(`Created ${home.group.name}. You are its Group Admin.`);
  }

  async function createGroupInvite() {
    if (!accessToken || !groupHome) {
      setStatus("Select a Group before creating an Invite.");
      return;
    }

    try {
      const invite = await createInvite({ baseUrl: apiBaseUrl, accessToken, groupId: groupHome.group.id });
      setInviteToken(invite.token);
      setStatus(`Created Invite for ${groupHome.group.name}.`);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Could not create Invite.");
    }
  }

  async function acceptGroupInvite() {
    if (!accessToken) {
      setStatus("Sign in before accepting an Invite.");
      return;
    }

    try {
      const home = await acceptInvite({ baseUrl: apiBaseUrl, accessToken, token: inviteToken.trim() });
      const groups = await listGroups({ baseUrl: apiBaseUrl, accessToken });
      setGroupHome(home);
      setSelectedJump(home.recentJumps[0] ?? null);
      setMemberships(groups.memberships);
      setInviteToken("");
      setStatus(`Joined ${home.group.name}.`);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Could not accept Invite.");
    }
  }

  async function selectGroup(token: string, groupId: string) {
    const home = await getGroupHome({ baseUrl: apiBaseUrl, accessToken: token, groupId });
    setGroupHome(home);
    setSelectedJump(home.recentJumps[0] ?? null);
    setIdea(null);
    setPlannedJump(null);
  }

  async function startNewSeason() {
    if (!accessToken || !groupHome) {
      setStatus("Select a Group before starting a Season.");
      return;
    }

    const home = await startSeason({ baseUrl: apiBaseUrl, accessToken, groupId: groupHome.group.id });
    setGroupHome(home);
    setSelectedJump(home.recentJumps[0] ?? null);
    setStatus(`Started an Active Season for ${home.group.name}. You are the Season Commissioner.`);
  }

  async function captureIdea() {
    if (!accessToken || !groupHome) {
      setStatus("Select a Group before creating an Idea.");
      return;
    }

    const captured = await createIdea({
      baseUrl: apiBaseUrl,
      accessToken,
      groupId: groupHome.group.id,
      source,
      destination,
      food,
    });
    setIdea(captured);
    setPlannedJump(null);
    setStatus(`Captured Idea: ${captured.food} from ${captured.source} at ${captured.destination}.`);
  }

  async function planJump(offSeason = false) {
    if (!accessToken || !idea) {
      setStatus("Create an Idea before creating a Planned Jump.");
      return;
    }

    const planned = await createPlannedJump({ baseUrl: apiBaseUrl, accessToken, ideaId: idea.id, offSeason });
    setPlannedJump(planned);
    setIdea(null);
    setSource("");
    setDestination("");
    setFood("");
    setStatus(
      planned.offSeason
        ? `Created an Off-Season Jump for ${planned.food}.`
        : `Created a Season-linked Planned Jump for ${planned.food}.`,
    );
  }

  async function startJudging(jump: PerformedJumpView) {
    if (!accessToken) {
      setStatus("Sign in to judge jumps.");
      return;
    }
    if (jump.performer.id === profile?.player.id) {
      setStatus("You cannot judge your own jump.");
      return;
    }
    setIsJudging(true);
    setGestureScores({ difficulty: 5, transgression: 5, creativity: 5, presentation: 5 });
    setSelectedJump(jump);
    setStatus("Use gestures or buttons to set scores. Swipe up/down on each factor.");
  }

  function updateGestureScore(factor: "difficulty" | "transgression" | "creativity" | "presentation", delta: number) {
    if (!gestureScores) return;
    const newValue = Math.max(0, Math.min(10, gestureScores[factor] + delta));
    setGestureScores({ ...gestureScores, [factor]: newValue });
  }

  function clearGestureScores() {
    setGestureScores({ difficulty: 5, transgression: 5, creativity: 5, presentation: 5 });
    setStatus("Gesture scores cleared. Start fresh.");
  }

  async function submitGestureJudgment() {
    if (!accessToken || !selectedJump || !gestureScores) {
      setStatus("Cannot submit judgment without scores.");
      return;
    }

    try {
      const judgment = await submitJudgment({
        baseUrl: apiBaseUrl,
        accessToken,
        jumpId: selectedJump.jump.id,
        ...gestureScores,
      });
      setPendingJudgment(judgment);
      setIsJudging(false);
      setGestureScores(null);
      setStatus(`Judgment submitted for ${selectedJump.jump.food}.`);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Could not submit judgment.");
    }
  }

  function cancelJudging() {
    setIsJudging(false);
    setGestureScores(null);
    setStatus("Judging cancelled.");
  }

  function createScorePanResponder(factor: "difficulty" | "transgression" | "creativity" | "presentation") {
    let lastY = 0;
    return PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: (_, gestureState) => Math.abs(gestureState.dy) > 10,
      onPanResponderGrant: (_, gestureState) => {
        lastY = gestureState.y0;
      },
      onPanResponderMove: (_, gestureState) => {
        const deltaY = lastY - gestureState.moveY;
        if (Math.abs(deltaY) > 50) {
          updateGestureScore(factor, deltaY > 0 ? 1 : -1);
          lastY = gestureState.moveY;
        }
      },
      onPanResponderRelease: () => {},
    });
  }

  const difficultyPan = createScorePanResponder("difficulty");
  const transgressionPan = createScorePanResponder("transgression");
  const creativityPan = createScorePanResponder("creativity");
  const presentationPan = createScorePanResponder("presentation");

  return (
    <SafeAreaView style={styles.screen}>
      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.card}>
        <Text style={styles.title}>Supperjumpin</Text>
        <Text style={styles.body}>{status}</Text>
        <TextInput
          autoCapitalize="none"
          keyboardType="email-address"
          onChangeText={setEmail}
          placeholder="player@example.com"
          style={styles.input}
          value={email}
        />
        <Button onPress={sendMagicLink} title="Send magic link" />
        {profile ? (
          <Text style={styles.body}>
            Account {profile.account.id} is attached to Player {profile.player.id}.
          </Text>
        ) : null}
        {profile ? (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Create Group</Text>
            <TextInput
              onChangeText={setGroupName}
              placeholder="Breakfast Crew"
              style={styles.input}
              value={groupName}
            />
            <Button onPress={createNewGroup} title="Create Group" />
          </View>
        ) : null}
        {memberships.length > 0 ? (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Your Groups</Text>
            {memberships.map((entry) => (
              <Button
                key={entry.group.id}
                onPress={() => accessToken && selectGroup(accessToken, entry.group.id)}
                title={`${entry.group.name} (${entry.membership.role})`}
              />
            ))}
          </View>
        ) : null}
        {profile ? (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Invites</Text>
            <Button onPress={createGroupInvite} title="Create Invite for selected Group" />
            <TextInput
              autoCapitalize="none"
              onChangeText={setInviteToken}
              placeholder="Invite token"
              style={styles.input}
              value={inviteToken}
            />
            <Button onPress={acceptGroupInvite} title="Accept Invite" />
          </View>
        ) : null}
        {groupHome ? (
          <View style={styles.groupHome}>
            <Text style={styles.groupName}>{groupHome.group.name}</Text>
            <Text style={styles.body}>Your Group Membership: {groupHome.membership.role}</Text>
            <Text style={styles.body}>
              Active Season: {groupHome.activeSeason ? groupHome.activeSeason.status : "None"}
            </Text>
            {groupHome.activeSeason ? (
              <Text style={styles.body}>
                Season Commissioner: {groupHome.activeSeason.commissionerPlayerId}
              </Text>
            ) : (
              <Button onPress={startNewSeason} title="Start Season" />
            )}
            <Text style={styles.body}>Recent Jumps: {groupHome.recentJumps.length}</Text>
            <Text style={styles.body}>Standings: {groupHome.standings.length}</Text>
            {groupHome.recentJumps.length > 0 ? (
              <View style={styles.section}>
                <Text style={styles.sectionTitle}>Recent Performed Jumps</Text>
                {groupHome.recentJumps.map((performedJump: PerformedJumpView) => (
                  <Button
                    key={performedJump.jump.id}
                    onPress={() => setSelectedJump(performedJump)}
                    title={`${performedJump.jump.food} at ${performedJump.jump.destination}`}
                  />
                ))}
              </View>
            ) : null}
            {selectedJump ? (
              <View style={styles.jumpCard}>
                <Text style={styles.sectionTitle}>Jump Detail</Text>
                <Text style={styles.body}>Source: {selectedJump.jump.source}</Text>
                <Text style={styles.body}>Destination: {selectedJump.jump.destination}</Text>
                <Text style={styles.body}>Food: {selectedJump.jump.food}</Text>
                <Text style={styles.body}>Performer: {selectedJump.performer.displayName}</Text>
                <Text style={styles.body}>Caption: {selectedJump.evidence.caption}</Text>
                <Text style={styles.body}>Media object: {selectedJump.evidence.mediaObjectKey}</Text>
                {pendingJudgment && selectedJump.jump.id === pendingJudgment.jumpId ? (
                  <View style={styles.judgmentSummary}>
                    <Text style={styles.sectionTitle}>Your Judgment</Text>
                    <Text style={styles.body}>Difficulty: {pendingJudgment.difficulty}</Text>
                    <Text style={styles.body}>Transgression: {pendingJudgment.transgression}</Text>
                    <Text style={styles.body}>Creativity: {pendingJudgment.creativity}</Text>
                    <Text style={styles.body}>Presentation: {pendingJudgment.presentation}</Text>
                  </View>
                ) : null}
                {!isJudging && selectedJump.performer.id !== profile?.player.id ? (
                  <Button onPress={() => startJudging(selectedJump)} title="Judge this Jump" />
                ) : null}
                {selectedJump.performer.id === profile?.player.id ? (
                  <Text style={styles.body}>Your own jump - cannot judge.</Text>
                ) : null}
              </View>
            ) : null}
            {isJudging && gestureScores ? (
              <View style={styles.judgingCard}>
                <Text style={styles.sectionTitle}>Quick-Judge</Text>
                <Text style={styles.body}>Swipe up/down on each factor to adjust score (0-10)</Text>
                <View style={styles.scoreRow} {...difficultyPan.panHandlers} accessibilityLabel="Difficulty score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Difficulty:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.difficulty} out of 10`}>{gestureScores.difficulty}</Text>
                  <Button onPress={() => updateGestureScore("difficulty", -1)} title="-" accessibilityLabel="Decrease difficulty score by 1" />
                  <Button onPress={() => updateGestureScore("difficulty", 1)} title="+" accessibilityLabel="Increase difficulty score by 1" />
                </View>
                <View style={styles.scoreRow} {...transgressionPan.panHandlers} accessibilityLabel="Transgression score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Transgression:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.transgression} out of 10`}>{gestureScores.transgression}</Text>
                  <Button onPress={() => updateGestureScore("transgression", -1)} title="-" accessibilityLabel="Decrease transgression score by 1" />
                  <Button onPress={() => updateGestureScore("transgression", 1)} title="+" accessibilityLabel="Increase transgression score by 1" />
                </View>
                <View style={styles.scoreRow} {...creativityPan.panHandlers} accessibilityLabel="Creativity score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Creativity:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.creativity} out of 10`}>{gestureScores.creativity}</Text>
                  <Button onPress={() => updateGestureScore("creativity", -1)} title="-" accessibilityLabel="Decrease creativity score by 1" />
                  <Button onPress={() => updateGestureScore("creativity", 1)} title="+" accessibilityLabel="Increase creativity score by 1" />
                </View>
                <View style={styles.scoreRow} {...presentationPan.panHandlers} accessibilityLabel="Presentation score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Presentation:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.presentation} out of 10`}>{gestureScores.presentation}</Text>
                  <Button onPress={() => updateGestureScore("presentation", -1)} title="-" accessibilityLabel="Decrease presentation score by 1" />
                  <Button onPress={() => updateGestureScore("presentation", 1)} title="+" accessibilityLabel="Increase presentation score by 1" />
                </View>
                <View style={styles.judgmentActions}>
                  <Button onPress={clearGestureScores} title="Clear" color="#c1673a" />
                  <Button onPress={cancelJudging} title="Cancel" color="#888" />
                  <Button onPress={submitGestureJudgment} title="Submit Judgment" color="#2f241d" />
                </View>
              </View>
            ) : null}
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>Create Idea</Text>
              <TextInput onChangeText={setSource} placeholder="Source" style={styles.input} value={source} />
              <TextInput
                onChangeText={setDestination}
                placeholder="Destination"
                style={styles.input}
                value={destination}
              />
              <TextInput onChangeText={setFood} placeholder="Food" style={styles.input} value={food} />
              <Button onPress={captureIdea} title="Capture Idea" />
              {idea ? (
                <View style={styles.jumpCard}>
                  <Text style={styles.body}>Idea: {idea.food}</Text>
                  <Button onPress={() => planJump(false)} title="Create Planned Jump" />
                  <Button onPress={() => planJump(true)} title="Create Off-Season Jump" />
                </View>
              ) : null}
              {plannedJump ? (
                <Text style={styles.body}>
                  Planned Jump: {plannedJump.source} to {plannedJump.destination} with {plannedJump.food} (
                  {plannedJump.offSeason ? "Off-Season Jump" : "Season-linked"})
                </Text>
              ) : null}
            </View>
          </View>
        ) : null}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
  content: {
    flexGrow: 1,
    justifyContent: "center",
    padding: 24,
  },
  card: {
    backgroundColor: "#fffaf2",
    borderColor: "#2f241d",
    borderRadius: 24,
    borderWidth: 2,
    gap: 16,
    padding: 24,
  },
  title: {
    color: "#2f241d",
    fontSize: 36,
    fontWeight: "900",
  },
  body: {
    color: "#4d3b31",
    fontSize: 16,
  },
  section: {
    gap: 8,
  },
  sectionTitle: {
    color: "#2f241d",
    fontSize: 18,
    fontWeight: "800",
  },
  groupHome: {
    backgroundColor: "#f7efe2",
    borderRadius: 16,
    gap: 8,
    padding: 16,
  },
  groupName: {
    color: "#2f241d",
    fontSize: 24,
    fontWeight: "900",
  },
  jumpCard: {
    backgroundColor: "#fffaf2",
    borderColor: "#c1673a",
    borderRadius: 12,
    borderWidth: 1,
    gap: 8,
    padding: 12,
  },
  judgingCard: {
    backgroundColor: "#e8f4e8",
    borderColor: "#2f7d2f",
    borderRadius: 12,
    borderWidth: 2,
    gap: 12,
    padding: 16,
  },
  scoreRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
    paddingVertical: 12,
    paddingHorizontal: 8,
    borderRadius: 8,
    backgroundColor: "rgba(255, 255, 255, 0.5)",
  },
  scoreLabel: {
    color: "#2f241d",
    fontSize: 14,
    fontWeight: "600",
    flex: 1,
  },
  scoreValue: {
    color: "#2f7d2f",
    fontSize: 20,
    fontWeight: "900",
    width: 32,
    textAlign: "center",
  },
  judgmentActions: {
    flexDirection: "row",
    gap: 8,
    justifyContent: "flex-end",
  },
  judgmentSummary: {
    backgroundColor: "#e8f4e8",
    borderColor: "#2f7d2f",
    borderRadius: 8,
    borderWidth: 1,
    gap: 4,
    padding: 12,
    marginTop: 8,
  },
  input: {
    backgroundColor: "white",
    borderColor: "#c1673a",
    borderRadius: 12,
    borderWidth: 1,
    fontSize: 16,
    padding: 12,
  },
});
