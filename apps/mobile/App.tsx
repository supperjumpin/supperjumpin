import { useCallback, useState } from "react";
import { SafeAreaView, StyleSheet, StatusBar } from "react-native";
import FeedScreen from "./FeedScreen";
import JumpDetailScreen from "./JumpDetailScreen";

type Screen =
  | { name: "feed" }
  | { name: "detail"; jumpId: string };

export default function App() {
  const [screen, setScreen] = useState<Screen>({ name: "feed" });

  const handleNavigateDetail = useCallback((jumpId: string) => {
    setScreen({ name: "detail", jumpId });
  }, []);

  const handleBack = useCallback(() => {
    setScreen({ name: "feed" });
  }, []);

  return (
    <SafeAreaView style={styles.screen}>
      <StatusBar barStyle="dark-content" backgroundColor="#f7efe2" />
      {screen.name === "feed" ? (
        <FeedScreen onNavigateDetail={handleNavigateDetail} />
      ) : (
        <JumpDetailScreen
          jumpId={screen.jumpId}
          onBack={handleBack}
          onBrowseFeed={handleBack}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
});
