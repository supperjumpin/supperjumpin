import React, { useState, useCallback, useEffect } from "react";
import {
  View,
  Text,
  FlatList,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
  TouchableOpacity,
  Image,
} from "react-native";
import { getPublicFeed } from "@supperjumpin/api-client";

const API_BASE = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const SUPABASE_URL = process.env.EXPO_PUBLIC_SUPABASE_URL ?? "";
const STORAGE_BUCKET = "evidence";

function mediaUrl(key: string): string | null {
  if (!key || !SUPABASE_URL) return null;
  return `${SUPABASE_URL}/storage/v1/object/public/${STORAGE_BUCKET}/${key}`;
}

function formatTimeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = now - then;
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function formatScore(avg: number, count: number): string {
  return `★ ${avg.toFixed(1)} (${count})`;
}

function formatCountdown(endsAt: string): string {
  const remaining = new Date(endsAt).getTime() - Date.now();
  if (remaining <= 0) return "0m 0s";
  const mins = Math.floor(remaining / 60000);
  const secs = Math.floor((remaining % 60000) / 1000);
  return `${mins}m ${secs}s`;
}

function formatGraceCountdown(endsAt: string): string {
  const countdown = formatCountdown(endsAt);
  if (countdown === "0m 0s") return "Judging now available";
  return `Judging opens in ${countdown}`;
}

function evidenceAltText(caption: string): string {
  return caption || "Evidence photo";
}

interface JumpCardProps {
  id: string;
  performerName: string;
  source: string;
  destination: string;
  food: string;
  caption: string;
  mediaObjectKey: string;
  runningAverage: number;
  judgmentCount: number;
  status: string;
  gracePeriodExpiresAt: string;
  createdAt: string;
  viewerContext?: {
    canJudge: boolean;
    reason?: string | null;
    gracePeriodEndsAt?: string | null;
    hasJudged: boolean;
  };
  onPress: (id: string) => void;
}

function JumpCard({
  id,
  performerName,
  source,
  destination,
  food,
  caption,
  mediaObjectKey,
  runningAverage,
  judgmentCount,
  status,
  gracePeriodExpiresAt,
  createdAt,
  viewerContext,
  onPress,
}: JumpCardProps) {
  const canJudge = viewerContext?.canJudge ?? true;
  const judgeReason = viewerContext?.reason;

  let judgeLabel = "Judge";
  let judgeAccessibility = "Judge this Jump";
  if (judgeReason === "self-judging") {
    judgeLabel = "Your Jump";
    judgeAccessibility = "You performed this jump. You cannot judge your own entry.";
  } else if (judgeReason === "grace-period") {
    judgeLabel = formatGraceCountdown(gracePeriodExpiresAt);
    judgeAccessibility = `Judging opens in ${formatCountdown(gracePeriodExpiresAt)}. Not yet available.`;
  } else if (judgeReason === "already-judged") {
    judgeLabel = "You judged this";
    judgeAccessibility = "You already judged this jump. Score submitted.";
  } else if (!canJudge) {
    judgeLabel = "Not available";
  }

  const isGracePeriod = judgeReason === "grace-period";
  const isAlreadyJudged = judgeReason === "already-judged";
  const isSelfJudging = judgeReason === "self-judging";

  return (
    <TouchableOpacity
      style={styles.card}
      onPress={() => onPress(id)}
      accessibilityRole="button"
      accessibilityLabel={`${performerName} jumped ${food} from ${source} to ${destination}. Score: ${runningAverage.toFixed(1)} out of 4 from ${judgmentCount} judgments. ${formatTimeAgo(createdAt)} ago.`}
      accessibilityHint="Opens jump detail"
    >
      <View style={styles.cardHeader}>
        {mediaUrl(mediaObjectKey) ? (
          <Image
            source={{ uri: mediaUrl(mediaObjectKey) as string }}
            style={styles.mediaImage}
            accessibilityLabel={evidenceAltText(caption)}
            resizeMode="cover"
          />
        ) : (
          <View style={styles.mediaPlaceholder}>
            <Text style={styles.mediaPlaceholderText}>📸</Text>
          </View>
        )}
        <View style={styles.cardMeta}>
          <Text
            style={styles.scoreBadge}
            accessibilityLabel={`Running average ${runningAverage.toFixed(1)} out of 4 from ${judgmentCount} judgments`}
          >
            {formatScore(runningAverage, judgmentCount)}
          </Text>
        </View>
      </View>

      <View style={styles.cardBody}>
        <Text style={styles.route} accessibilityLabel={`From ${source} to ${destination}`}>
          {source} → {destination}
        </Text>
        <Text style={styles.food} accessibilityLabel={food}>
          {food}
        </Text>
        <Text style={styles.caption} numberOfLines={2} accessibilityLabel={caption}>
          {caption}
        </Text>
        <View style={styles.cardFooter}>
          <Text style={styles.performer} accessibilityLabel={`${performerName}`}>
            {performerName}
          </Text>
          <Text style={styles.timestamp}>
            {formatTimeAgo(createdAt)}
          </Text>
        </View>
      </View>

      <View style={styles.cardActions}>
        <View
          style={[
            styles.judgeButton,
            (isAlreadyJudged || isSelfJudging) && styles.judgeButtonDone,
            isGracePeriod && styles.judgeButtonGrace,
          ]}
          accessibilityLabel={judgeAccessibility}
          accessibilityRole="button"
        >
          <Text
            style={[
              styles.judgeButtonText,
              (isAlreadyJudged || isSelfJudging) && styles.judgeButtonTextDone,
              isGracePeriod && styles.judgeButtonTextGrace,
            ]}
            numberOfLines={1}
            accessibilityLabel={
              isGracePeriod
                ? `Judging opens in ${formatCountdown(gracePeriodExpiresAt)}. Not yet available.`
                : judgeAccessibility
            }
          >
            {judgeLabel}
          </Text>
        </View>
      </View>
    </TouchableOpacity>
  );
}

interface FeedScreenProps {
  onNavigateDetail: (jumpId: string) => void;
}

export default function FeedScreen({ onNavigateDetail }: FeedScreenProps) {
  const [jumps, setJumps] = useState<any[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchFeed = useCallback(
    async (cursor?: string, append = false) => {
      try {
        const data = await getPublicFeed({
          baseUrl: API_BASE,
          cursor,
          limit: 20,
        });
        if (append) {
          setJumps((prev) => [...prev, ...data.jumps]);
        } else {
          setJumps(data.jumps);
        }
        setNextCursor(data.nextCursor);
        setError(null);
      } catch (e: any) {
        setError(e.message ?? "Could not load jumps");
      }
    },
    []
  );

  useEffect(() => {
    (async () => {
      setLoading(true);
      await fetchFeed();
      setLoading(false);
    })();
  }, [fetchFeed]);

  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchFeed();
    setRefreshing(false);
  }, [fetchFeed]);

  const handleLoadMore = useCallback(async () => {
    if (!nextCursor || loadingMore) return;
    setLoadingMore(true);
    await fetchFeed(nextCursor, true);
    setLoadingMore(false);
  }, [nextCursor, loadingMore, fetchFeed]);

  const renderCard = useCallback(
    ({ item }: { item: any }) => (
      <JumpCard
        id={item.id}
        performerName={item.performerName}
        source={item.source}
        destination={item.destination}
        food={item.food}
        caption={item.caption}
        mediaObjectKey={item.mediaObjectKey}
        runningAverage={item.runningAverage}
        judgmentCount={item.judgmentCount}
        status={item.status}
        gracePeriodExpiresAt={item.gracePeriodExpiresAt}
        createdAt={item.createdAt}
        viewerContext={item.viewerContext}
        onPress={onNavigateDetail}
      />
    ),
    [onNavigateDetail]
  );

  if (loading) {
    return (
      <View style={styles.centerContainer} accessible accessibilityLabel="Loading jumps">
        <ActivityIndicator size="large" color="#c1673a" />
        <Text style={styles.loadingText}>Loading jumps...</Text>
      </View>
    );
  }

  if (error && jumps.length === 0) {
    return (
      <View style={styles.centerContainer} accessible accessibilityLabel="Could not load jumps">
        <Text style={styles.errorText} accessibilityLabel="Could not load jumps. Network error.">
          Could not load jumps. Network error.
        </Text>
          <TouchableOpacity
            style={styles.retryButton}
            onPress={handleRefresh}
            accessibilityRole="button"
            accessibilityLabel="Retry loading jumps"
            accessibilityHint="Double tap to retry"
          >
            <Text style={styles.retryButtonText}>Retry</Text>
          </TouchableOpacity>
        </View>
      );
  }

  return (
    <View style={styles.screen}>
      <View style={styles.header}>
        <Text style={styles.title}>Supperjumpin</Text>
      </View>
      <FlatList
        testID="feed-list"
        data={jumps}
        renderItem={renderCard}
        keyExtractor={(item: any) => item.id}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={handleRefresh}
            accessibilityLabel="Pull to refresh"
          />
        }
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
        ListEmptyComponent={
          <View style={styles.centerContainer} accessible accessibilityLabel="No jumps yet. The feed is empty." accessibilityHint="Double tap to refresh">
            <Text style={styles.emptyText}>No jumps yet. The feed is empty.</Text>
          </View>
        }
        ListFooterComponent={
          loadingMore ? (
            <View style={styles.footerLoader}>
              <ActivityIndicator size="small" color="#c1673a" />
            </View>
          ) : null
        }
        contentContainerStyle={jumps.length === 0 ? styles.emptyList : undefined}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
  },
  header: {
    paddingHorizontal: 16,
    paddingTop: 16,
    paddingBottom: 8,
  },
  title: {
    color: "#2f241d",
    fontSize: 32,
    fontWeight: "900",
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
  retryButton: {
    backgroundColor: "#c1673a",
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 12,
    minHeight: 48,
    justifyContent: "center",
  },
  retryButtonText: {
    color: "#fffaf2",
    fontSize: 16,
    fontWeight: "700",
  },
  emptyText: {
    color: "#4d3b31",
    fontSize: 18,
    fontWeight: "600",
    textAlign: "center",
  },
  emptyList: {
    flexGrow: 1,
  },
  card: {
    backgroundColor: "#fffaf2",
    borderColor: "#2f241d",
    borderRadius: 24,
    borderWidth: 2,
    marginHorizontal: 16,
    marginVertical: 6,
    overflow: "hidden",
  },
  cardHeader: {
    flexDirection: "row",
    alignItems: "flex-start",
  },
  mediaPlaceholder: {
    width: "100%",
    height: 180,
    backgroundColor: "#e8ddd2",
    justifyContent: "center",
    alignItems: "center",
  },
  mediaImage: {
    width: "100%",
    height: 180,
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
  },
  mediaPlaceholderText: {
    fontSize: 48,
  },
  cardMeta: {
    position: "absolute",
    top: 8,
    right: 8,
  },
  scoreBadge: {
    backgroundColor: "#2f241d",
    color: "#fffaf2",
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
    fontSize: 14,
    fontWeight: "800",
    overflow: "hidden",
  },
  cardBody: {
    padding: 16,
    gap: 6,
  },
  route: {
    color: "#c1673a",
    fontSize: 14,
    fontWeight: "700",
  },
  food: {
    color: "#2f241d",
    fontSize: 20,
    fontWeight: "900",
  },
  caption: {
    color: "#4d3b31",
    fontSize: 14,
    lineHeight: 20,
  },
  cardFooter: {
    flexDirection: "row",
    justifyContent: "space-between",
    marginTop: 4,
  },
  performer: {
    color: "#2f241d",
    fontSize: 14,
    fontWeight: "600",
  },
  timestamp: {
    color: "#888",
    fontSize: 12,
  },
  cardActions: {
    paddingHorizontal: 16,
    paddingBottom: 16,
  },
  judgeButton: {
    backgroundColor: "#c1673a",
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 12,
    minHeight: 44,
    justifyContent: "center",
    alignItems: "center",
  },
  judgeButtonDone: {
    backgroundColor: "#e8ddd2",
  },
  judgeButtonGrace: {
    backgroundColor: "#ddd",
  },
  judgeButtonText: {
    color: "#fffaf2",
    fontSize: 14,
    fontWeight: "700",
  },
  judgeButtonTextDone: {
    color: "#888",
  },
  judgeButtonTextGrace: {
    color: "#888",
    fontSize: 12,
  },
  footerLoader: {
    paddingVertical: 20,
    alignItems: "center",
  },
});
