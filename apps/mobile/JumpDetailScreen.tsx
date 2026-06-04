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

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const SUPABASE_URL = process.env.EXPO_PUBLIC_SUPABASE_URL ?? "";
const STORAGE_BUCKET = "evidence";

function mediaUrl(key: string): string | null {
  if (!key || !SUPABASE_URL) return null;
  return `${SUPABASE_URL}/storage/v1/object/public/${STORAGE_BUCKET}/${key}`;
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatGraceCountdown(endsAt: string): string {
  const remaining = new Date(endsAt).getTime() - Date.now();
  if (remaining <= 0) return "Judging now available";
  const mins = Math.floor(remaining / 60000);
  const secs = Math.floor((remaining % 60000) / 1000);
  return `Judging opens in ${mins}m ${secs}s`;
}

function evidenceAltText(detail: {
  caption: string;
  food: string;
  source: string;
  destination: string;
}): string {
  if (detail.caption) return detail.caption;
  return `Evidence photo for ${detail.food} from ${detail.source} to ${detail.destination}`;
}

interface JumpDetailScreenProps {
  jumpId: string;
  onBack: () => void;
  onBrowseFeed: () => void;
}

export default function JumpDetailScreen({
  jumpId,
  onBack,
  onBrowseFeed,
}: JumpDetailScreenProps) {
  const [detail, setDetail] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isTombstone, setIsTombstone] = useState(false);

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const data = await getJumpDetail({ baseUrl: API_BASE, jumpId });
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
  }, [jumpId]);

  if (loading) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#c1673a" />
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.centerContainer}>
        <Text style={styles.errorText}>{error}</Text>
        <TouchableOpacity style={styles.button} onPress={onBack}>
          <Text style={styles.buttonText}>Back to Feed</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (isTombstone) {
    return (
      <View style={styles.screen}>
        <TouchableOpacity style={styles.backButton} onPress={onBack}>
          <Text style={styles.backText}>← Back</Text>
        </TouchableOpacity>
        <View style={styles.centerContainer}>
          <Text style={styles.tombstoneMessage}>
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

  const vc = detail.viewerContext;
  let judgeState: string;
  if (!vc || vc.canJudge) {
    judgeState = "Judge this Jump";
  } else if (vc.reason === "self-judging") {
    judgeState = "You can't judge your own Jump";
  } else if (vc.reason === "grace-period") {
    judgeState = formatGraceCountdown(detail.gracePeriodExpiresAt);
  } else if (vc.reason === "already-judged") {
    judgeState = "You already judged this jump. Score submitted.";
  } else {
    judgeState = "Not available";
  }

  const canJudge = vc ? vc.canJudge : true;

  return (
    <View style={styles.screen}>
      <TouchableOpacity style={styles.backButton} onPress={onBack}>
        <Text style={styles.backText}>← Back</Text>
      </TouchableOpacity>

      <ScrollView contentContainerStyle={styles.content}>
        {/* Evidence photo */}
        {mediaUrl(detail.mediaObjectKey) ? (
          <Image
            source={{ uri: mediaUrl(detail.mediaObjectKey) as string }}
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
          <Text style={styles.performerName}>
            {detail.performerName}
          </Text>
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
          accessibilityLabel={judgeState}
        >
          <Text
            style={[
              styles.judgeStateText,
              canJudge ? styles.judgeStateActive : styles.judgeStateInactive,
            ]}
          >
            {judgeState}
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
