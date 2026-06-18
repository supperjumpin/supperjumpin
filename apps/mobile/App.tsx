import { useCallback, useEffect, useRef, useState, type Dispatch, type SetStateAction } from "react";
import { StyleSheet, StatusBar } from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { getMe, updateDisplayName } from "@supperjumpin/api-client";
import { createAppFlow, type FlowSnapshot, type JumpDraft } from "./flow";
import FeedScreen from "./FeedScreen";
import JumpDetailScreen from "./JumpDetailScreen";
import CreateJumpScreen from "./CreateJumpScreen";
import DisplayNameSetupScreen from "./DisplayNameSetupScreen";

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const DEV_AUTH_TOKEN = process.env.EXPO_PUBLIC_DEV_AUTH_TOKEN ?? "dev-token";

export default function App() {
  const flowRef = useRef(createAppFlow());
  const [flowState, setFlowState] = useState<FlowSnapshot>(() => flowRef.current.getSnapshot());

  const syncFlow = useCallback(() => {
    setFlowState(flowRef.current.getSnapshot());
  }, []);

  const { screen, session, player, showDisplayNameSetup, draft } = flowState;

  useEffect(() => {
    if (!session) return;
    let cancelled = false;
    (async () => {
      try {
        const me = await getMe({ baseUrl: API_BASE, accessToken: session.access_token });
        if (cancelled) return;
        flowRef.current.resolvePlayer({ id: me.player.id, displayName: me.player.displayName ?? "" });
        syncFlow();
      } catch {
        if (!cancelled) {
          flowRef.current.rejectPlayer();
          syncFlow();
        }
      }
    })();
    return () => { cancelled = true; };
  }, [session, syncFlow]);

  const handleSignIn = useCallback(() => {
    flowRef.current.signIn({ access_token: DEV_AUTH_TOKEN, user: { id: "signed-in-player" } });
    syncFlow();
  }, [syncFlow]);

  const handleRequestAuth = useCallback(() => {
    flowRef.current.navigateToCreate();
    syncFlow();
    flowRef.current.signIn({ access_token: DEV_AUTH_TOKEN, user: { id: "signed-in-player" } });
    syncFlow();
  }, [syncFlow]);

  const handleDisplayNameSet = useCallback(async (displayName: string) => {
    const s = flowRef.current.getSnapshot().session;
    if (!s) return;
    const result = await updateDisplayName({ baseUrl: API_BASE, accessToken: s.access_token, displayName });
    flowRef.current.setDisplayName(result.player.displayName);
    syncFlow();
  }, [syncFlow]);

  const handleNavigateDetail = useCallback((jumpId: string) => {
    flowRef.current.navigateToDetail(jumpId);
    syncFlow();
  }, [syncFlow]);

  const handleNavigateCreate = useCallback(() => {
    flowRef.current.navigateToCreate();
    syncFlow();
  }, [syncFlow]);

  const handleBack = useCallback(() => {
    flowRef.current.navigateBack();
    syncFlow();
  }, [syncFlow]);

  const handleJumpDraftSubmitSuccess = useCallback(() => {
    flowRef.current.submitDraft();
    syncFlow();
  }, [syncFlow]);

  const handleDraftChange: Dispatch<SetStateAction<JumpDraft>> = useCallback((action) => {
    const current = flowRef.current.getSnapshot().draft;
    const next = typeof action === 'function' ? action(current) : action;
    flowRef.current.changeDraft(next);
    syncFlow();
  }, [syncFlow]);

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
          draft={draft}
          onDraftChange={handleDraftChange}
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
            onSignIn={handleRequestAuth}
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
