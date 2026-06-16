import { useCallback, useEffect, useRef, useState } from "react";
import { SafeAreaView, StyleSheet, StatusBar } from "react-native";
import { GoTrueClient } from "@supabase/auth-js";
import { getMe, updateDisplayName } from "@supperjumpin/api-client";
import FeedScreen from "./FeedScreen";
import JumpDetailScreen from "./JumpDetailScreen";
import CreateJumpScreen from "./CreateJumpScreen";
import DisplayNameSetupScreen from "./DisplayNameSetupScreen";

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
  | { name: "detail"; jumpId: string }
  | { name: "create" };

export default function App() {
  const [screen, setScreen] = useState<Screen>({ name: "feed" });
  const [session, setSession] = useState<Session | null>(null);
  const [player, setPlayer] = useState<Player | null>(null);
  const [showDisplayNameSetup, setShowDisplayNameSetup] = useState(false);
  const [client] = useState(() => SUPABASE_URL ? new GoTrueClient({ url: SUPABASE_URL, autoRefreshToken: false, persistSession: false }) : null);

  const pendingCreateRef = useRef(false);

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

  // Keep session in sync after OAuth redirects while the app stays mounted.
  useEffect(() => {
    if (!client) return;

    const { data } = client.onAuthStateChange((_event, authSession) => {
      if (authSession) {
        setSession({ access_token: authSession.access_token, user: { id: authSession.user.id } });
      } else {
        setSession(null);
      }
    });

    return () => {
      data.subscription.unsubscribe();
    };
  }, [client]);

  // When session changes, fetch player profile
  useEffect(() => {
    if (!session) {
      setPlayer(null);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const me = await getMe({ baseUrl: API_BASE, accessToken: session.access_token });
        if (cancelled) return;
        setPlayer({ id: me.player.id, displayName: me.player.displayName ?? "" });

        // If user tapped "+ Jump" before auth, complete the flow
        if (pendingCreateRef.current) {
          if (me.player.displayName) {
            pendingCreateRef.current = false;
            setScreen({ name: "create" });
          } else {
            setShowDisplayNameSetup(true);
          }
        }
      } catch {
        if (!cancelled) setSession(null);
      }
    })();
    return () => { cancelled = true; };
  }, [session]);

  const handleSignIn = useCallback(async () => {
    if (!client) return;
    const { error } = await client.signInWithOAuth({ provider: "google" });
    if (error) throw error;
  }, [client]);

  const handleRequestAuth = useCallback(() => {
    pendingCreateRef.current = true;
    void handleSignIn();
  }, [handleSignIn]);

  const handleDisplayNameSet = useCallback(async (displayName: string) => {
    if (!session) return;
    const result = await updateDisplayName({ baseUrl: API_BASE, accessToken: session.access_token, displayName });
    setPlayer({ id: result.player.id, displayName: result.player.displayName });
    setShowDisplayNameSetup(false);
    pendingCreateRef.current = false;
    setScreen({ name: "create" });
  }, [session]);

  const handleNavigateDetail = useCallback((jumpId: string) => {
    setScreen({ name: "detail", jumpId });
  }, []);

  const handleNavigateCreate = useCallback(() => {
    if (session) {
      if (!player) {
        pendingCreateRef.current = true;
        return;
      }
      if (!player.displayName) {
        pendingCreateRef.current = true;
        setShowDisplayNameSetup(true);
        return;
      }
    }
    setScreen({ name: "create" });
  }, [player, session]);

  const handleBack = useCallback(() => {
    setScreen({ name: "feed" });
  }, []);

  const content = (
    <SafeAreaView style={styles.screen}>
      <StatusBar barStyle="dark-content" backgroundColor="#f7efe2" />
      {screen.name === "feed" ? (
        <FeedScreen
          onNavigateDetail={handleNavigateDetail}
          onNavigateCreate={handleNavigateCreate}
          onRequestAuth={handleRequestAuth}
          session={session}
        />
      ) : screen.name === "create" ? (
        <CreateJumpScreen onBack={handleBack} />
      ) : (() => {
        const s = screen as { name: "detail"; jumpId: string };
        return (
          <JumpDetailScreen
            jumpId={s.jumpId}
            onBack={handleBack}
            onBrowseFeed={handleBack}
            session={session}
          />
        );
      })()}
    </SafeAreaView>
  );

  // Display name setup overlay — shown when first-time auth user taps "+ Jump"
  // before confirming their display name
  if (showDisplayNameSetup) {
    return (
      <DisplayNameSetupScreen onSubmit={handleDisplayNameSet} />
    );
  }

  return content;
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
});
