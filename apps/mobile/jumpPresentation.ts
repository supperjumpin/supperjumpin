export interface JudgeDisplay {
  label: string;
  accessibilityLabel: string;
  reason: "self-judging" | "grace-period" | "already-judged" | "can-judge" | "not-available";
}

export function mediaUrl(key: string, baseUrl?: string): string | null {
  if (!key || !baseUrl) return null;
  return `${baseUrl.replace(/\/$/, "")}/${key.replace(/^\//, "")}`;
}

export function evidenceAltText(jump: {
  caption: string;
  food?: string;
  source?: string;
  destination?: string;
}): string {
  if (jump.caption) return jump.caption;
  if (jump.food && jump.source && jump.destination) {
    return `Evidence photo for ${jump.food} from ${jump.source} to ${jump.destination}`;
  }
  return "Evidence photo";
}

export function formatScore(avg: number, count: number): string {
  return `★ ${avg.toFixed(1)} (${count})`;
}

export function formatCountdown(endsAt: string, nowMs: number = Date.now()): string {
  const remaining = new Date(endsAt).getTime() - nowMs;
  if (remaining <= 0) return "0m 0s";
  const mins = Math.floor(remaining / 60000);
  const secs = Math.floor((remaining % 60000) / 1000);
  return `${mins}m ${secs}s`;
}

export function formatGracePeriodLabel(endsAt: string, nowMs: number = Date.now()): string {
  const countdown = formatCountdown(endsAt, nowMs);
  if (countdown === "0m 0s") return "Judging now available";
  return `Judging opens in ${countdown}`;
}

export function judgeLabel(
  viewerContext: { canJudge: boolean; hasJudged?: boolean; reason?: string | null } | undefined | null,
  gracePeriodExpiresAt?: string | null,
  nowMs: number = Date.now()
): JudgeDisplay {
  if (viewerContext?.reason === "self-judging") {
    return {
      label: "You can't judge your own Jump",
      accessibilityLabel: "You performed this jump. You cannot judge your own entry.",
      reason: "self-judging",
    };
  }

  if (viewerContext?.reason === "grace-period" && gracePeriodExpiresAt) {
    const label = formatGracePeriodLabel(gracePeriodExpiresAt, nowMs);
    const countdown = formatCountdown(gracePeriodExpiresAt, nowMs);
    return {
      label,
      accessibilityLabel: `Judging opens in ${countdown}. Not yet available.`,
      reason: "grace-period",
    };
  }

  if (viewerContext?.reason === "already-judged") {
    return {
      label: "You already judged this jump. Score submitted.",
      accessibilityLabel: "You already judged this jump. Score submitted.",
      reason: "already-judged",
    };
  }

  if (viewerContext && !viewerContext.canJudge) {
    return {
      label: "Not available",
      accessibilityLabel: "Not available",
      reason: "not-available",
    };
  }

  return { label: "Judge this Jump", accessibilityLabel: "Judge this Jump", reason: "can-judge" };
}
