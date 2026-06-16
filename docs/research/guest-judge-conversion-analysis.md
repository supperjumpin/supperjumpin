# Guest Judge Conversion Analysis

## Executive Summary

The Guest Judge model in Supperjumpin contains a **material contradiction** between the product vision (auth gates Judging) and the domain model (Guest Judges may Judge without auth). The MVP roadmap resolves this by deferring the auth gate on Judging to v2, but this leaves a critical gap: **the only conversion nudge in v1 is "claim your history," which is insufficient to hit the 15% Guest-to-Player conversion target without additional mechanics.**

This document identifies the contradiction, maps the actual flow, benchmarks conversion rates, and proposes specific, low-friction conversion mechanics that fit within MVP scope.

---

## 1. The Auth Model Contradiction

### What the Documents Say

| Source | Claim |
|--------|-------|
| **Product Vision** (01-product-vision.md, line 43-49) | "Hits a soft auth gate when they try to Judge — one-tap social login... Auth gates contribution, not consumption." |
| **CONTEXT.md** (line 68-70) | "Guest Judge: A visitor who submits a Judgment without creating an Account... In v1, all Judging is available to Guest Judges; a soft auth cap may be introduced in v2 based on data." |
| **MVP Roadmap** (04-mvp-roadmap.md, line 59, 64) | "Auth gates contribution (Judging, posting), not consumption... Any authenticated Player or Guest Judge may Judge." |
| **ADR-0019** (line 11) | "Judging on the public feed is open: any Player may Judge a Jump they did not perform." (silent on Guest Judges) |

### The Contradiction

The product vision describes an auth-gated Judging experience: you browse freely, but to Judge you must authenticate. This is the "soft auth gate" described in the first 5-minute experience.

The domain model and MVP roadmap describe the opposite: Guest Judges may Judge without any auth. The auth gate applies only to **posting** a Jump, not to Judging.

### Resolution in Current Docs

The MVP roadmap resolves the tension by **deferring the auth gate on Judging to v2** ("soft auth cap may be introduced in v2 based on data"). In v1, Judging is fully open to Guest Judges. The auth wall appears only when:

1. **Posting a Jump** — the CTA surfaces after the first Judgment, and tapping it requires auth.
2. **Viewing your own Judgment history** — to "claim your history," you must create an Account.

This means the actual first auth gate is **posting**, not Judging. The product vision's step 3 ("hits a soft auth gate when they try to Judge") is **aspirational for v2, not accurate for v1**.

### Gap: No Auth Gate on Judging Means No Forced Conversion Point

By allowing Guest Judges to Judge without auth, the product removes the most natural conversion moment: the point where a user has invested effort (rendering a Judgment) and has a clear reason to convert (to save that Judgment). The only remaining conversion triggers are:

- **Post-Judgment CTA** — "Want to post your own Jump? Create an Account."
- **History claim** — "View your Judgment history" (requires auth to persist across sessions).

Both are weaker than an auth gate at the moment of Judging because they occur **after** the user has already received the core value (Judging) for free.

---

## 2. The Actual Guest Judge Flow (v1)

```
1. Unauthenticated visitor opens app
   → Lands on public feed (no login gate)

2. Browses 2-3 Jumps
   → No auth required

3. Taps "Judge" on a Jump
   → NO AUTH GATE (Guest Judge may Judge freely)
   → Judgment stored by device/session ID

4. Submits Judgment
   → Running average updates
   → CTA surfaces: "Post your own Jump" OR "View your Judgment history"

5. Taps CTA
   → AUTH GATE APPEARS (one-tap social login + email magic link option)
   → If auth succeeds: Guest Judgments are migrated to new Account
```

**First auth wall: Step 5 (posting or viewing history), NOT Step 3 (Judging).**

---

## 3. Conversion Mechanics at Small Scale (100-500 Users)

The MVP roadmap explicitly defers push notifications, Missions, Levels, and progression systems. What remains?

### Mechanics That Fit Within MVP Scope

| Mechanic | How It Works | Friction | Expected Lift |
|----------|--------------|----------|---------------|
| **"Claim your history"** | Guest Judges see a persistent banner: "You've Judged 3 Jumps. Create an Account to save your history." | Low | Baseline — mentioned in docs as the only nudge |
| **The Open (monthly competition)** | Guest Judges see a prompt: "Create an Account to compete in the Open and see your Standings." | Low | Moderate — gives a reason to convert beyond history |
| **Post-Judgment CTA** | After submitting a Judgment, surface: "Want to post your own Jump? It takes 30 seconds." | Low | Moderate — strikes while engagement is high |
| **Social proof in feed** | Show "X Players are competing this month" or "Y Judgments submitted today" | Zero | Low-Moderate — FOMO without requiring auth |
| **Judgment receipt share** | After Judging, offer: "Share your verdict" with a card showing what you scored | Zero | Moderate — shareable artifact creates identity investment |
| **Session persistence warning** | "Your Judgments are saved on this device. Switch devices or clear data and you'll lose them." | Zero | Low — fear of loss |
| **Limited Guest Judgments (soft cap)** | Allow N free Judgments, then require auth (the v2 "soft auth cap" brought forward) | Medium | High — forced conversion point, but risks drop-off |

### Mechanics That Require Post-MVP (Explicitly Deferred)

| Mechanic | Deferred To | Why It Doesn't Fit MVP |
|----------|-------------|------------------------|
| Push notifications | Post-MVP | Requires notification infrastructure and opt-in |
| Missions / Levels | Post-MVP | Progression systems require retention baseline |
| Email drip campaigns | Post-MVP | No email collection without auth; no email infra in MVP |
| Referral rewards | Post-MVP | Requires invite flow and tracking |
| Group Seasons | v2 | Requires Group infrastructure |

### What Works Without Push/Progression/Email

Based on research of apps at small scale:

1. **Duolingo model**: Let users experience core value (Judging) freely, then gate persistence/progress. Supperjumpin already does this — the gap is making the "loss of progress" salient.
2. **Pinterest model**: Browse freely, auth only when saving. Supperjumpin's equivalent is "auth when posting or viewing history."
3. **Wordle model**: No auth needed, but shareable artifacts create identity. Supperjumpin's Judgment receipt share is the closest equivalent.
4. **Headspace model**: Experience first, auth later to track progress. Supperjumpin could show "You've Judged X Jumps" as a progress indicator even for Guests.

### Recommended MVP Conversion Stack

To maximize conversion without deferred features, Supperjumpin should implement:

1. **Persistent Guest banner** — "You've Judged 3 Jumps on this device. Create an Account to save them." (always visible, updates with count).
2. **Post-Judgment CTA** — Immediately after submitting a Judgment, offer: "Post your own Jump" or "See how your scores compare" (both require auth).
3. **The Open prompt** — If the Guest has Judged ≥1 Jump in the current month, show: "You're eligible for this month's Open. Create an Account to compete."
4. **Judgment receipt share** — Let Guests share their verdict as a card. The share text includes: "I judged this Jump on Supperjumpin. Create an Account to judge your own." (viral loop with conversion hook).
5. **Session loss warning** — On app open (if Guest has Judgments), show: "Your 5 Judgments are saved on this device only."

---

## 4. Benchmarks: Is 15% Realistic?

### Industry Baselines

| Context | Conversion Rate | Source |
|---------|----------------|--------|
| Anonymous visitor → registered (news publishers) | 0.5–2% monthly | Playwire/INMA |
| Anonymous visitor → registered (gaming/sports) | 1–2%+ | Playwire |
| SaaS visitor → signup | 2–5% | Heap, Klipfolio |
| Social login vs. email signup | +8.2 percentage points | Heap |
| Social login improvement | 20–40% lift | CIAM Compass |
| Freemium app → paid | 2–5% | Kirro/RevenueCat |
| Trial-to-paid (well-designed) | 40–60% | RevenueCat |

### Key Insight

The 15% target is **not a visitor-to-signup rate** — it's a **Guest Judge-to-Player rate**. This is a much narrower funnel:

```
All Visitors
  → Some browse
    → Some Judge as Guest (already engaged)
      → Some create Account (the 15% target)
```

Guests who have already Judged are **higher-intent** than raw visitors. They have:
- Invested time in the app
- Understood the core mechanic
- Generated data (Judgments) they may want to keep

### Comparable Benchmarks for Engaged Users

| Context | Rate | Notes |
|---------|------|-------|
| Duolingo: lesson completers → account creation | ~15-25% | Estimated from public data; users who complete a lesson are highly engaged |
| Pinterest: savers → account creation | ~10-15% | Users who try to save a pin are engaged |
| Headspace: meditation completers → account | ~8-12% | Users who complete a session |
| E-commerce: guest checkout → account creation | 5-10% | Post-purchase prompts |

### Verdict on 15%

**Achievable but aggressive.** For users who have already Judged 2+ Jumps, 15% is plausible if:
- The conversion prompt is immediate and contextual (post-Judgment).
- There is a clear value proposition (history + Open competition).
- Social login is one-tap (Apple + Google).

**Without any conversion-optimized flows, 15% is unrealistic.** The baseline for engaged-but-unauthenticated users converting is 5-10%. To hit 15%, Supperjumpin needs:

1. **At least one forced decision point** (e.g., "You've used your 3 free Judgments. Create an Account to continue.").
2. **Or strong identity investment** (e.g., shareable Judgment receipts that create social identity).
3. **Or loss aversion** (e.g., "Your Judgments will be lost if you don't create an Account").

The current docs mention only "claim your history" — this alone is unlikely to hit 15%.

---

## 5. The First Auth Gate Experience

### What the Docs Say

| Source | Detail |
|--------|--------|
| Product Vision | "One-tap social login; no email verification on first tap" |
| MVP Roadmap | "One-tap social login (no email verification on first tap)" |
| ADR-0006 | Hosted auth is deferred until the local MVP is playable end-to-end; local development uses a static bearer token while preserving internal Accounts. |

### What's Missing

**No explicit provider list.** Hosted auth is deferred, so provider choice remains unresolved. When social login is introduced, iOS requirements still shape the provider set:

| Provider | Likely Included? | Rationale |
|----------|------------------|-----------|
| **Apple Sign-In** | Yes | App Store-required if any third-party social login is offered |
| **Google** | Likely | Large reach and low-friction consumer login |
| **Email magic link** | Yes | Mentioned explicitly in ADR-0006; no password required |
| Facebook | Maybe | Declining usage; adds complexity |
| Microsoft/GitHub | No | B2B/developer-focused, not relevant |

### Recommended Auth Gate UX

```
[Screen: Post-Judgment CTA or History Claim]

"Save your Judgments and compete in the Open"

[Sign in with Apple]  ← primary on iOS
[Sign in with Google] ← primary on Android
[Continue with Email] ← sends magic link

"By continuing, you agree to our Terms and Privacy Policy.
We never post without your permission."
```

**Key principles:**
- No password creation (magic links + social login only).
- No email verification step (social login bypasses this; magic links are self-verifying).
- One tap to initiate, zero typing for social login.
- Guest Judgments migrate automatically on auth success.

---

## 6. Specific Gaps and Recommendations

### Gap 1: No Forced Conversion Point

**Problem**: Guests can Judge indefinitely without ever seeing an auth wall. The only conversion moments are optional CTAs.

**Recommendation**: Introduce a **soft Judgment cap** in v1 (not v2): allow 3 Guest Judgments, then require auth to continue. This is not a hard paywall — it's a "save your progress" prompt. This aligns with the product vision's "soft auth gate" while preserving the Guest Judge model.

### Gap 2: Weak Value Proposition for Conversion

**Problem**: "Claim your history" is the only reason to convert. For a Guest who has Judged 1-2 Jumps, history is not compelling.

**Recommendation**: Add **The Open eligibility** as a conversion hook. After the first Judgment, show: "Your Judgments count toward this month's Open. Create an Account to see your Standings." Competition is a stronger motivator than data persistence.

### Gap 3: No Identity Investment

**Problem**: Guests have no social identity in the app. They are invisible.

**Recommendation**: Add **shareable Judgment receipts**. After Judging, let Guests share a card: "I scored this Jump: Commitment 3/4, Transgression 4/4..." The share includes a deep link to the Jump. This creates identity investment ("I rendered a verdict") and viral distribution simultaneously.

### Gap 4: No Loss Aversion

**Problem**: Guests don't know their Judgments are ephemeral.

**Recommendation**: Add a **persistent, non-intrusive banner**: "You've Judged 3 Jumps on this device. Create an Account to save them permanently." Update the count after each Judgment. Make the loss salient without being aggressive.

### Gap 5: Auth Gate Timing Is Wrong

**Problem**: The product vision says auth gates Judging, but the implementation gates posting. This misalignment means the team may be optimizing the wrong funnel.

**Recommendation**: **Decide explicitly** whether v1 auth gates Judging or posting, and align all docs. Two options:

| Option | Auth Gates | Pros | Cons |
|--------|-----------|------|------|
| A: Vision-aligned | Judging | Higher conversion rate; matches product vision; creates clear funnel | Higher drop-off at first interaction; fewer total Judgments |
| B: Current model | Posting | More total Judgments; lower friction; aligns with CONTEXT.md | Lower conversion rate; no forced decision point |

**Suggested decision**: Option B (current model) for First Playable Loop, then test Option A (auth-gated Judging) in MVP if Guest-to-Player conversion is below 15%.

---

## 7. Summary Table: Contradictions and Gaps

| # | Issue | Source A | Source B | Severity |
|---|-------|----------|----------|----------|
| 1 | Auth gates Judging vs. Guest Judges allowed | Product Vision (01) | CONTEXT.md, MVP Roadmap (04) | **High** — fundamental flow mismatch |
| 2 | Only conversion nudge is "claim history" | CONTEXT.md | — | **High** — insufficient for 15% target |
| 3 | No explicit auth providers listed | Product Vision, ADR-0006 | — | Medium — implementation risk |
| 4 | No forced conversion point in v1 | MVP Roadmap | Product Vision | **High** — 15% unlikely without it |
| 5 | Guest Judges are invisible (no identity) | — | — | Medium — missed viral opportunity |
| 6 | No loss aversion for ephemeral Judgments | — | — | Medium — Guests don't know data is at risk |

---

## 8. Recommended Next Steps

1. **Align the auth model** — Update Product Vision or CONTEXT.md to resolve the contradiction. Decide: does auth gate Judging (vision) or posting (current implementation)?
2. **Add conversion mechanics to MVP scope** — The following are small enough to fit in v1 and necessary to hit 15%:
   - Persistent Guest banner with Judgment count.
   - Post-Judgment CTA with Open eligibility hook.
   - Shareable Judgment receipts.
3. **Specify auth providers** — Document Apple + Google + email magic link as the v1 auth stack.
4. **Define the soft auth cap** — If keeping Guest Judges, specify the cap (e.g., 3 Judgments) and the prompt text.
5. **Instrument conversion funnel** — Track: Guest opens → Guest browses → Guest Judges → Guest sees CTA → Guest taps CTA → Guest creates Account. Identify the biggest drop-off.

---

*Analysis conducted on 2026-05-30. Sources: CONTEXT.md, Product Vision (01), MVP Roadmap (04), ADR-0006, ADR-0019, ADR-0023, market research (issue #62), and industry conversion benchmarks.*
