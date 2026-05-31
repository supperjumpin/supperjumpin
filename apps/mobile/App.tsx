import { GoTrueClient } from "@supabase/auth-js";
import { useEffect, useState } from "react";
import { Button, SafeAreaView, ScrollView, StyleSheet, Text, TextInput, View, PanResponder } from "react-native";

import {
  acceptInvite,
  createGroup,
  createIdea,
  createInvite,
  createPlannedStunt,
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
  PerformedStuntView,
  Stunt,
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
  const [selectedStunt, setSelectedStunt] = useState<PerformedStuntView | null>(null);
  const [idea, setIdea] = useState<Stunt | null>(null);
  const [plannedStunt, setPlannedStunt] = useState<Stunt | null>(null);
  const [gestureScores, setGestureScores] = useState<{ difficulty: number; transgression: number; creativity: number; documentation: number } | null>(null);
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
    setSelectedStunt(home.recentStunts[0] ?? null);
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
      setSelectedStunt(home.recentStunts[0] ?? null);
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
    setSelectedStunt(home.recentStunts[0] ?? null);
    setIdea(null);
    setPlannedStunt(null);
  }

  async function startNewSeason() {
    if (!accessToken || !groupHome) {
      setStatus("Select a Group before starting a Season.");
      return;
    }

    const home = await startSeason({ baseUrl: apiBaseUrl, accessToken, groupId: groupHome.group.id });
    setGroupHome(home);
    setSelectedStunt(home.recentStunts[0] ?? null);
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
    setPlannedStunt(null);
    setStatus(`Captured Idea: ${captured.food} from ${captured.source} at ${captured.destination}.`);
  }

  async function planStunt(offSeason = false) {
    if (!accessToken || !idea) {
      setStatus("Create an Idea before creating a Planned Stunt.");
      return;
    }

    const planned = await createPlannedStunt({ baseUrl: apiBaseUrl, accessToken, ideaId: idea.id, offSeason });
    setPlannedStunt(planned);
    setIdea(null);
    setSource("");
    setDestination("");
    setFood("");
    setStatus(
      planned.offSeason
        ? `Created an Off-Season Stunt for ${planned.food}.`
        : `Created a Season-linked Planned Stunt for ${planned.food}.`,
    );
  }

  async function startJudging(stunt: PerformedStuntView) {
    if (!accessToken) {
      setStatus("Sign in to judge stunts.");
      return;
    }
    if (stunt.performer.id === profile?.player.id) {
      setStatus("You cannot judge your own stunt.");
      return;
    }
    setIsJudging(true);
    setGestureScores({ difficulty: 3, transgression: 3, creativity: 3, documentation: 3 });
    setSelectedStunt(stunt);
    setStatus("Use gestures or buttons to set scores. Swipe up/down on each factor.");
  }

  function updateGestureScore(factor: "difficulty" | "transgression" | "creativity" | "documentation", delta: number) {
    if (!gestureScores) return;
    const newValue = Math.max(1, Math.min(4, gestureScores[factor] + delta));
    setGestureScores({ ...gestureScores, [factor]: newValue });
  }

  function clearGestureScores() {
    setGestureScores({ difficulty: 3, transgression: 3, creativity: 3, documentation: 3 });
    setStatus("Gesture scores cleared. Start fresh.");
  }

  async function submitGestureJudgment() {
    if (!accessToken || !selectedStunt || !gestureScores) {
      setStatus("Cannot submit judgment without scores.");
      return;
    }

    try {
      const judgment = await submitJudgment({
        baseUrl: apiBaseUrl,
        accessToken,
        stuntId: selectedStunt.stunt.id,
        ...gestureScores,
      });
      setPendingJudgment(judgment);
      setIsJudging(false);
      setGestureScores(null);
      setStatus(`Judgment submitted for ${selectedStunt.stunt.food}.`);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Could not submit judgment.");
    }
  }

  function cancelJudging() {
    setIsJudging(false);
    setGestureScores(null);
    setStatus("Judging cancelled.");
  }

  function createScorePanResponder(factor: "difficulty" | "transgression" | "creativity" | "documentation") {
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
  const documentationPan = createScorePanResponder("documentation");

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
            <Text style={styles.body}>Recent Stunts: {groupHome.recentStunts.length}</Text>
            <Text style={styles.body}>Standings: {groupHome.standings.length}</Text>
            {groupHome.recentStunts.length > 0 ? (
              <View style={styles.section}>
                <Text style={styles.sectionTitle}>Recent Performed Stunts</Text>
                {groupHome.recentStunts.map((performedStunt: PerformedStuntView) => (
                  <Button
                    key={performedStunt.stunt.id}
                    onPress={() => setSelectedStunt(performedStunt)}
                    title={`${performedStunt.stunt.food} at ${performedStunt.stunt.destination}`}
                  />
                ))}
              </View>
            ) : null}
            {selectedStunt ? (
              <View style={styles.stuntCard}>
                <Text style={styles.sectionTitle}>Stunt Detail</Text>
                <Text style={styles.body}>Source: {selectedStunt.stunt.source}</Text>
                <Text style={styles.body}>Destination: {selectedStunt.stunt.destination}</Text>
                <Text style={styles.body}>Food: {selectedStunt.stunt.food}</Text>
                <Text style={styles.body}>Performer: {selectedStunt.performer.displayName}</Text>
                <Text style={styles.body}>Caption: {selectedStunt.evidence.caption}</Text>
                <Text style={styles.body}>Media object: {selectedStunt.evidence.mediaObjectKey}</Text>
                {pendingJudgment && selectedStunt.stunt.id === pendingJudgment.stuntId ? (
                  <View style={styles.judgmentSummary}>
                    <Text style={styles.sectionTitle}>Your Judgment</Text>
                    <Text style={styles.body}>Difficulty: {pendingJudgment.difficulty}</Text>
                    <Text style={styles.body}>Transgression: {pendingJudgment.transgression}</Text>
                    <Text style={styles.body}>Creativity: {pendingJudgment.creativity}</Text>
                    <Text style={styles.body}>Documentation: {pendingJudgment.documentation}</Text>
                  </View>
                ) : null}
                {!isJudging && selectedStunt.performer.id !== profile?.player.id ? (
                  <Button onPress={() => startJudging(selectedStunt)} title="Judge this Stunt" />
                ) : null}
                {selectedStunt.performer.id === profile?.player.id ? (
                  <Text style={styles.body}>Your own stunt - cannot judge.</Text>
                ) : null}
              </View>
            ) : null}
            {isJudging && gestureScores ? (
              <View style={styles.judgingCard}>
                <Text style={styles.sectionTitle}>Quick-Judge</Text>
                <Text style={styles.body}>Swipe up/down on each factor to adjust score (1-4)</Text>
                <View style={styles.scoreRow} {...difficultyPan.panHandlers} accessibilityLabel="Difficulty score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Difficulty:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.difficulty} out of 4`}>{gestureScores.difficulty}</Text>
                  <Button onPress={() => updateGestureScore("difficulty", -1)} title="-" accessibilityLabel="Decrease difficulty score by 1" />
                  <Button onPress={() => updateGestureScore("difficulty", 1)} title="+" accessibilityLabel="Increase difficulty score by 1" />
                </View>
                <View style={styles.scoreRow} {...transgressionPan.panHandlers} accessibilityLabel="Transgression score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Transgression:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.transgression} out of 4`}>{gestureScores.transgression}</Text>
                  <Button onPress={() => updateGestureScore("transgression", -1)} title="-" accessibilityLabel="Decrease transgression score by 1" />
                  <Button onPress={() => updateGestureScore("transgression", 1)} title="+" accessibilityLabel="Increase transgression score by 1" />
                </View>
                <View style={styles.scoreRow} {...creativityPan.panHandlers} accessibilityLabel="Creativity score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Creativity:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.creativity} out of 4`}>{gestureScores.creativity}</Text>
                  <Button onPress={() => updateGestureScore("creativity", -1)} title="-" accessibilityLabel="Decrease creativity score by 1" />
                  <Button onPress={() => updateGestureScore("creativity", 1)} title="+" accessibilityLabel="Increase creativity score by 1" />
                </View>
                <View style={styles.scoreRow} {...documentationPan.panHandlers} accessibilityLabel="Documentation score adjustment. Swipe up or down to change value.">
                  <Text style={styles.scoreLabel}>Documentation:</Text>
                  <Text style={styles.scoreValue} accessibilityLabel={`Current score: ${gestureScores.documentation} out of 4`}>{gestureScores.documentation}</Text>
                  <Button onPress={() => updateGestureScore("documentation", -1)} title="-" accessibilityLabel="Decrease documentation score by 1" />
                  <Button onPress={() => updateGestureScore("documentation", 1)} title="+" accessibilityLabel="Increase documentation score by 1" />
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
                <View style={styles.stuntCard}>
                  <Text style={styles.body}>Idea: {idea.food}</Text>
                  <Button onPress={() => planStunt(false)} title="Create Planned Stunt" />
                  <Button onPress={() => planStunt(true)} title="Create Off-Season Stunt" />
                </View>
              ) : null}
              {plannedStunt ? (
                <Text style={styles.body}>
                  Planned Stunt: {plannedStunt.source} to {plannedStunt.destination} with {plannedStunt.food} (
                  {plannedStunt.offSeason ? "Off-Season Stunt" : "Season-linked"})
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
  stuntCard: {
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
