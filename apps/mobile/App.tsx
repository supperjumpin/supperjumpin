import { useCallback, useEffect, useRef, useState } from "react";
import { StyleSheet, StatusBar } from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { getMe, updateDisplayName } from "@supperjumpin/api-client";
import FeedScreen from "./FeedScreen";
import JumpDetailScreen from "./JumpDetailScreen";
import CreateJumpScreen from "./CreateJumpScreen";
import DisplayNameSetupScreen from "./DisplayNameSetupScreen";

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const DEV_AUTH_TOKEN = process.env.EXPO_PUBLIC_DEV_AUTH_TOKEN ?? "dev-token";

interface Session {
  access_token: string;
  user: { id: string };
}

interface Player {
  id: string;
  displayName: string;
}

interface JumpDraft {
  source: string;
  destination: string;
  food: string;
  caption: string;
  mediaObjectKey: string;
}

const emptyJumpDraft = (): JumpDraft => ({
  source: '',
  destination: '',
  food: '',
  caption: '',
  mediaObjectKey: '',
});

type Screen =
  | { name: "feed" }
  | { name: "detail"; jumpId: string }
  | { name: "create" };

export default function App() {
  const [screen, setScreen] = useState<Screen>({ name: "feed" });
  const [session, setSession] = useState<Session | null>(null);
  const [player, setPlayer] = useState<Player | null>(null);
  const [showDisplayNameSetup, setShowDisplayNameSetup] = useState(false);
  const [jumpDraft, setJumpDraft] = useState<JumpDraft>(emptyJumpDraft);

  const pendingCreateRef = useRef(false);

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
    setSession({ access_token: DEV_AUTH_TOKEN, user: { id: "signed-in-player" } });
  }, []);

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

  const handleJumpDraftSubmitSuccess = useCallback(() => {
    setJumpDraft(emptyJumpDraft());
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
        <CreateJumpScreen
          session={session}
          onBack={handleBack}
          draft={jumpDraft}
          onDraftChange={setJumpDraft}
          onSubmitSuccess={handleJumpDraftSubmitSuccess}
        />
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

  const appContent = showDisplayNameSetup ? (
    <DisplayNameSetupScreen onSubmit={handleDisplayNameSet} />
  ) : content;

  return <SafeAreaProvider>{appContent}</SafeAreaProvider>;
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
});
