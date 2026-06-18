import {
  mediaUrl,
  evidenceAltText,
  formatScore,
  formatCountdown,
  formatGracePeriodLabel,
  judgeLabel,
} from "./jumpPresentation";

describe("mediaUrl", () => {
  test("returns null when key is empty", () => {
    expect(mediaUrl("", "https://media.example.com")).toBeNull();
  });

  test("returns null when baseUrl is empty", () => {
    expect(mediaUrl("photos/abc.jpg", "")).toBeNull();
  });

  test("returns null when baseUrl is undefined", () => {
    expect(mediaUrl("photos/abc.jpg")).toBeNull();
  });

  test("returns null when both arguments are empty", () => {
    expect(mediaUrl("", "")).toBeNull();
  });

  test("constructs URL from key and base URL", () => {
    expect(mediaUrl("photos/abc.jpg", "https://media.example.com")).toBe(
      "https://media.example.com/photos/abc.jpg"
    );
  });

  test("strips trailing slash from base URL", () => {
    expect(mediaUrl("photos/abc.jpg", "https://media.example.com/")).toBe(
      "https://media.example.com/photos/abc.jpg"
    );
  });

  test("strips leading slash from key", () => {
    expect(mediaUrl("/photos/abc.jpg", "https://media.example.com")).toBe(
      "https://media.example.com/photos/abc.jpg"
    );
  });
});

describe("evidenceAltText", () => {
  test("returns caption when available", () => {
    expect(
      evidenceAltText({
        caption: "Made it across the parking lot",
        food: "Crunchwrap",
        source: "Taco Bell",
        destination: "Olive Garden",
      })
    ).toBe("Made it across the parking lot");
  });

  test("returns descriptive text from food/route when no caption", () => {
    expect(
      evidenceAltText({
        caption: "",
        food: "Crunchwrap",
        source: "Taco Bell",
        destination: "Olive Garden",
      })
    ).toBe("Evidence photo for Crunchwrap from Taco Bell to Olive Garden");
  });

  test("returns fallback when no caption and no food/route", () => {
    expect(evidenceAltText({ caption: "" })).toBe("Evidence photo");
  });

  test("returns fallback when no caption and partial food/route", () => {
    expect(
      evidenceAltText({ caption: "", food: "Crunchwrap" })
    ).toBe("Evidence photo");
  });
});

describe("formatScore", () => {
  test("formats average and count", () => {
    expect(formatScore(3.5, 4)).toBe("★ 3.5 (4)");
  });

  test("formats single-digit average", () => {
    expect(formatScore(4.0, 10)).toBe("★ 4.0 (10)");
  });
});

describe("formatCountdown", () => {
  test("returns 0m 0s when time has passed", () => {
    expect(formatCountdown("2020-01-01T00:00:00Z", Date.now())).toBe("0m 0s");
  });

  test("formats minutes and seconds", () => {
    const now = Date.now();
    const endsAt = new Date(now + 3 * 60000 + 45000).toISOString();
    expect(formatCountdown(endsAt, now)).toBe("3m 45s");
  });

  test("uses Date.now() as default nowMs", () => {
    const future = new Date(Date.now() + 60000).toISOString();
    const result = formatCountdown(future);
    expect(result).toMatch(/^1m \d+s$/);
  });
});

describe("formatGracePeriodLabel", () => {
  test("returns Judging now available when countdown expired", () => {
    expect(
      formatGracePeriodLabel("2020-01-01T00:00:00Z", Date.now())
    ).toBe("Judging now available");
  });

  test("returns countdown label when time remaining", () => {
    const endsAt = new Date(Date.now() + 2 * 60000 + 30000).toISOString();
    expect(formatGracePeriodLabel(endsAt, Date.now())).toBe(
      "Judging opens in 2m 30s"
    );
  });
});

describe("judgeLabel", () => {
  test("returns can-judge when viewerContext is null", () => {
    const result = judgeLabel(null);
    expect(result.reason).toBe("can-judge");
    expect(result.label).toBe("Judge this Jump");
    expect(result.accessibilityLabel).toBe("Judge this Jump");
  });

  test("returns can-judge when viewerContext is undefined", () => {
    const result = judgeLabel(undefined);
    expect(result.reason).toBe("can-judge");
  });

  test("returns can-judge when canJudge is true", () => {
    const result = judgeLabel({ canJudge: true, hasJudged: false });
    expect(result.reason).toBe("can-judge");
    expect(result.label).toBe("Judge this Jump");
  });

  test("returns self-judging when reason is self-judging", () => {
    const result = judgeLabel({
      canJudge: false,
      hasJudged: false,
      reason: "self-judging",
    });
    expect(result.reason).toBe("self-judging");
    expect(result.label).toBe("You can't judge your own Jump");
    expect(result.accessibilityLabel).toBe(
      "You performed this jump. You cannot judge your own entry."
    );
  });

  test("returns grace-period with countdown", () => {
    const endsAt = new Date(Date.now() + 5 * 60000).toISOString();
    const result = judgeLabel(
      { canJudge: false, hasJudged: false, reason: "grace-period" },
      endsAt,
      Date.now()
    );
    expect(result.reason).toBe("grace-period");
    expect(result.label).toBe("Judging opens in 5m 0s");
    expect(result.accessibilityLabel).toBe(
      "Judging opens in 5m 0s. Not yet available."
    );
  });

  test("returns already-judged", () => {
    const result = judgeLabel({
      canJudge: false,
      hasJudged: true,
      reason: "already-judged",
    });
    expect(result.reason).toBe("already-judged");
    expect(result.label).toBe("You already judged this jump. Score submitted.");
    expect(result.accessibilityLabel).toBe(
      "You already judged this jump. Score submitted."
    );
  });

  test("returns not-available when canJudge is false without specific reason", () => {
    const result = judgeLabel({ canJudge: false, hasJudged: false });
    expect(result.reason).toBe("not-available");
    expect(result.label).toBe("Not available");
  });
});
