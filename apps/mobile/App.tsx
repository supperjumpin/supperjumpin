import { createClient } from "@supabase/supabase-js";
import { useEffect, useState } from "react";
import { Button, SafeAreaView, StyleSheet, Text, TextInput, View } from "react-native";

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
} from "@supperjumpin/api-client";
import type { GroupHomeResponse, GroupMembershipSummary, MeResponse, Stunt } from "@supperjumpin/api-client";

const supabase = createClient(
  process.env.EXPO_PUBLIC_SUPABASE_URL ?? "",
  process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY ?? "",
);

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
  const [idea, setIdea] = useState<Stunt | null>(null);
  const [plannedStunt, setPlannedStunt] = useState<Stunt | null>(null);

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

  return (
    <SafeAreaView style={styles.screen}>
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
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
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
  input: {
    backgroundColor: "white",
    borderColor: "#c1673a",
    borderRadius: 12,
    borderWidth: 1,
    fontSize: 16,
    padding: 12,
  },
});
