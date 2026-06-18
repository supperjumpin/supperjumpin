import React, { useEffect, useState } from "react";
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Image,
} from "react-native";
import { getJumpDetail } from "@supperjumpin/api-client";
import { mediaUrl, evidenceAltText, judgeLabel } from "./jumpPresentation";
import { useLiveNow } from "./useLiveNow";

export { mediaUrl } from "./jumpPresentation";

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const MEDIA_BASE_URL = process.env.EXPO_PUBLIC_MEDIA_BASE_URL ?? "";

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

interface JumpDetailScreenProps {
  jumpId: string;
  onBack: () => void;
  onBrowseFeed: () => void;
  session?: { access_token: string; user: { id: string } } | null;
}

export default function JumpDetailScreen({
  jumpId,
  onBack,
  onBrowseFeed,
  session,
}: JumpDetailScreenProps) {
  const [detail, setDetail] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isTombstone, setIsTombstone] = useState(false);

  const isGraceActive = detail?.viewerContext?.reason === "grace-period";
  const nowMs = useLiveNow(isGraceActive);

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const data = await getJumpDetail({
          baseUrl: API_BASE,
          accessToken: session?.access_token,
          jumpId,
        });
        if (data && data.status === "Removed Jump") {
          setIsTombstone(true);
          setDetail(data);
        } else {
          setDetail(data);
        }
        setError(null);
      } catch (e: any) {
        setError(e.message ?? "Could not load jump detail");
      }
      setLoading(false);
    })();
  }, [jumpId, session?.access_token]);

  if (loading) {
    return (
      <View style={styles.centerContainer} accessible accessibilityLabel="Loading jump detail">
        <ActivityIndicator size="large" color="#c1673a" />
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.centerContainer} accessible accessibilityLabel="Could not load jump detail">
        <Text style={styles.errorText} accessibilityLabel={error}>{error}</Text>
        <TouchableOpacity
          style={styles.button}
          onPress={onBack}
          accessibilityRole="button"
          accessibilityLabel="Back to Feed"
        >
          <Text style={styles.buttonText}>Back to Feed</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (isTombstone) {
    return (
      <View style={styles.screen}>
        <TouchableOpacity
          style={styles.backButton}
          onPress={onBack}
          accessibilityRole="button"
          accessibilityLabel="Back to Feed"
        >
          <Text style={styles.backText}>← Back</Text>
        </TouchableOpacity>
        <View style={styles.centerContainer} accessible accessibilityLabel="Jump no longer available">
          <Text style={styles.tombstoneMessage} accessibilityLabel="This Jump is no longer available">
            This Jump is no longer available
          </Text>
          <TouchableOpacity
            style={styles.button}
            onPress={onBrowseFeed}
            accessibilityRole="button"
            accessibilityLabel="Browse Feed"
          >
            <Text style={styles.buttonText}>Browse Feed</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  const judgeDisplay = judgeLabel(
    detail.viewerContext,
    detail.gracePeriodExpiresAt,
    nowMs
  );

  const canJudge = judgeDisplay.reason === "can-judge";

  return (
    <View style={styles.screen}>
      <TouchableOpacity
        style={styles.backButton}
        onPress={onBack}
        accessibilityRole="button"
        accessibilityLabel="Back to Feed"
      >
        <Text style={styles.backText}>← Back</Text>
      </TouchableOpacity>

      <ScrollView contentContainerStyle={styles.content}>
        {/* Evidence photo */}
        {mediaUrl(detail.mediaObjectKey, MEDIA_BASE_URL) ? (
          <Image
            source={{ uri: mediaUrl(detail.mediaObjectKey, MEDIA_BASE_URL) as string }}
            style={styles.heroImageReal}
            accessibilityLabel={evidenceAltText(detail)}
            resizeMode="cover"
          />
        ) : (
          <View style={styles.heroImage}>
            <Text style={styles.heroImageText}>📸</Text>
          </View>
        )}

        {/* Performer + timestamp */}
        <View style={styles.performerRow}>
          <TouchableOpacity
            style={styles.performerStub}
            disabled
            accessibilityRole="button"
            accessibilityState={{ disabled: true }}
            accessibilityLabel={`${detail.performerName} profile coming soon`}
            accessibilityHint="Player profiles are not available yet."
          >
            <Text style={styles.performerName}>{detail.performerName}</Text>
          </TouchableOpacity>
          <Text style={styles.timestamp}>
            {formatDate(detail.createdAt)}
          </Text>
        </View>

        {/* Route */}
        <Text style={styles.route}>
          {detail.source} → {detail.destination}
        </Text>
        <Text style={styles.food}>{detail.food}</Text>

        {/* Caption */}
        <Text style={styles.caption}>{detail.caption}</Text>

        {/* Scores */}
        <View style={styles.scoresCard}>
          <Text style={styles.scoreLabel}>
            Running Average
          </Text>
          <Text
            style={styles.scoreValue}
            accessibilityLabel={`Running average ${detail.runningAverage.toFixed(1)} out of 4 from ${detail.judgmentCount} judgments`}
          >
            {detail.runningAverage.toFixed(1)}
          </Text>
          <Text style={styles.scoreSubtext}>
            from {detail.judgmentCount} judgment{detail.judgmentCount !== 1 ? "s" : ""}
          </Text>
        </View>

        {/* Judge action area */}
        <View
          style={[
            styles.judgeArea,
            canJudge ? styles.judgeAreaActive : styles.judgeAreaInactive,
          ]}
          accessible
          accessibilityLabel={judgeDisplay.label}
        >
          <Text
            style={[
              styles.judgeStateText,
              canJudge ? styles.judgeStateActive : styles.judgeStateInactive,
            ]}
          >
            {judgeDisplay.label}
          </Text>
        </View>

        {/* Disputes */}
        {detail.disputes && detail.disputes.length > 0 && (
          <View style={styles.disputesSection}>
            <Text style={styles.disputesLabel}>
              {detail.disputes.length} dispute{detail.disputes.length !== 1 ? "s" : ""}
            </Text>
          </View>
        )}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
  centerContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    padding: 32,
  },
  loadingText: {
    color: "#4d3b31",
    fontSize: 16,
    marginTop: 12,
  },
  errorText: {
    color: "#c1673a",
    fontSize: 18,
    fontWeight: "600",
    marginBottom: 16,
  },
  backButton: {
    paddingHorizontal: 16,
    paddingVertical: 12,
    minHeight: 44,
    justifyContent: "center",
  },
  backText: {
    color: "#c1673a",
    fontSize: 16,
    fontWeight: "700",
  },
  content: {
    padding: 16,
    gap: 12,
  },
  heroImage: {
    width: "100%",
    height: 250,
    backgroundColor: "#e8ddd2",
    borderRadius: 16,
    justifyContent: "center",
    alignItems: "center",
  },
  heroImageReal: {
    width: "100%",
    height: 250,
    borderRadius: 16,
  },
  heroImageText: {
    fontSize: 64,
  },
  performerRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
  },
  performerStub: {
    minHeight: 44,
    justifyContent: "center",
  },
  performerName: {
    color: "#2f241d",
    fontSize: 18,
    fontWeight: "800",
  },
  timestamp: {
    color: "#888",
    fontSize: 14,
  },
  route: {
    color: "#c1673a",
    fontSize: 16,
    fontWeight: "700",
  },
  food: {
    color: "#2f241d",
    fontSize: 28,
    fontWeight: "900",
  },
  caption: {
    color: "#4d3b31",
    fontSize: 16,
    lineHeight: 24,
  },
  scoresCard: {
    backgroundColor: "#2f241d",
    borderRadius: 16,
    padding: 20,
    alignItems: "center",
  },
  scoreLabel: {
    color: "#c1673a",
    fontSize: 14,
    fontWeight: "600",
    textTransform: "uppercase",
    letterSpacing: 1,
  },
  scoreValue: {
    color: "#fffaf2",
    fontSize: 48,
    fontWeight: "900",
  },
  scoreSubtext: {
    color: "#e8ddd2",
    fontSize: 14,
  },
  judgeArea: {
    borderRadius: 16,
    padding: 16,
    alignItems: "center",
    minHeight: 56,
    justifyContent: "center",
  },
  judgeAreaActive: {
    backgroundColor: "#c1673a",
  },
  judgeAreaInactive: {
    backgroundColor: "#e8ddd2",
  },
  judgeStateText: {
    fontSize: 18,
    fontWeight: "700",
  },
  judgeStateActive: {
    color: "#fffaf2",
  },
  judgeStateInactive: {
    color: "#888",
  },
  tombstoneMessage: {
    color: "#4d3b31",
    fontSize: 20,
    fontWeight: "700",
    textAlign: "center",
    marginBottom: 24,
  },
  button: {
    backgroundColor: "#c1673a",
    paddingHorizontal: 24,
    paddingVertical: 14,
    borderRadius: 12,
    minHeight: 48,
    justifyContent: "center",
  },
  buttonText: {
    color: "#fffaf2",
    fontSize: 16,
    fontWeight: "700",
    textAlign: "center",
  },
  disputesSection: {
    padding: 12,
    backgroundColor: "#fffaf2",
    borderRadius: 12,
    borderWidth: 1,
    borderColor: "#c1673a",
  },
  disputesLabel: {
    color: "#c1673a",
    fontSize: 14,
    fontWeight: "600",
    textAlign: "center",
  },
});
