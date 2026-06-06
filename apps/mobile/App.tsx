import { useCallback, useEffect, useState } from "react";
import { SafeAreaView, StyleSheet, StatusBar } from "react-native";
import { GoTrueClient } from "@supabase/auth-js";
import { getMe, updateDisplayName } from "@supperjumpin/api-client";
import FeedScreen from "./FeedScreen";
import JumpDetailScreen from "./JumpDetailScreen";
import AuthGate from "./AuthGate";

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const SUPABASE_URL = process.env.EXPO_PUBLIC_SUPABASE_URL ?? "";

interface Session {
  access_token: string;
  user: { id: string };
}

interface Player {
  id: string;
  displayName: string;
}

type Screen =
  | { name: "feed" }
  | { name: "detail"; jumpId: string };

export default function App() {
  const [screen, setScreen] = useState<Screen>({ name: "feed" });
  const [session, setSession] = useState<Session | null>(null);
  const [player, setPlayer] = useState<Player | null>(null);
  const [client] = useState(() => SUPABASE_URL ? new GoTrueClient({ url: SUPABASE_URL, autoRefreshToken: false, persistSession: false }) : null);

  // On mount, check for existing session
  useEffect(() => {
    (async () => {
      if (!client) return;
      const { data } = await client.getSession();
      if (data?.session) {
        setSession({ access_token: data.session.access_token, user: { id: data.session.user.id } });
      }
    })();
  }, [client]);

  // When session changes, fetch player profile
  useEffect(() => {
    if (!session) {
      setPlayer(null);
      return;
    }
    (async () => {
      try {
        const me = await getMe({ baseUrl: API_BASE, accessToken: session.access_token });
        setPlayer({ id: me.player.id, displayName: me.player.displayName ?? "" });
      } catch {
        // Session invalid — reset
        setSession(null);
      }
    })();
  }, [session]);

  const handleSignIn = useCallback(async () => {
    if (!client) return;
    const { error } = await client.signInWithOAuth({ provider: "google" });
    if (error) throw error;
    // OAuth redirects the browser; session arrives via URL callback
  }, [client]);

  const handleDisplayNameSet = useCallback(async (displayName: string) => {
    if (!session) return;
    const result = await updateDisplayName({ baseUrl: API_BASE, accessToken: session.access_token, displayName });
    setPlayer({ id: result.player.id, displayName: result.player.displayName });
  }, [session]);

  const handleNavigateDetail = useCallback((jumpId: string) => {
    setScreen({ name: "detail", jumpId });
  }, []);

  const handleBack = useCallback(() => {
    setScreen({ name: "feed" });
  }, []);

  const content = (
    <SafeAreaView style={styles.screen}>
      <StatusBar barStyle="dark-content" backgroundColor="#f7efe2" />
      {screen.name === "feed" ? (
        <FeedScreen onNavigateDetail={handleNavigateDetail} session={session} />
      ) : (
        <JumpDetailScreen
          jumpId={screen.jumpId}
          onBack={handleBack}
          onBrowseFeed={handleBack}
        />
      )}
    </SafeAreaView>
  );

  return (
    <AuthGate
      session={session}
      player={player}
      onSignIn={handleSignIn}
      onDisplayNameSet={handleDisplayNameSet}
    >
      {content}
    </AuthGate>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
});