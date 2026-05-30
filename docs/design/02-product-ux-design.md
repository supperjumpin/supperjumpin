# Product/UX Design

_Part of the [Supperjumpin Design Package](./README.md). Feeds: [Backend/Data Architecture](./03-backend-data-architecture.md)._

## 1. Core Loop

The Supperjumpin player journey is designed around a **low-friction entry, high-commitment creation** asymmetry. Browsing and Judging require seconds; posting a Jump requires real-world effort. This shapes every screen and interaction decision.

### 1.1 Primary Loop (Authenticated Player)

```
Feed Screen → Tap Jump Card → Jump Detail Screen → Tap "Judge" →
Judging Screen (tap-to-select 4 factors) → Judgment Receipt →
Jump Detail Screen (updated scores) → Scroll Feed / Tap "Post Jump" →
Create Jump Screen (photo + Caption + Source/Destination/Food) →
Submit → Feed Screen (new Jump appears after Author Grace Period)
```

### 1.2 Guest Loop (Unauthenticated Visitor)

```
Feed Screen → Tap Jump Card → Jump Detail Screen → Tap "Judge" →
Judging Screen (tap-to-select 4 factors) → Judgment Receipt →
Jump Detail Screen → CTA: "Post your own Jump" OR "View your Judging history" →
Auth Gate Screen (one-tap social login) → Create Jump Screen
```

**Key design principle:** The first auth wall appears at contribution (posting or viewing history), not at consumption or Judging. Guest Judges may submit up to 5 Judgments before encountering a soft auth cap (server-configurable; default 5 for v1).

### 1.3 Loop States and Transitions

| State | Player Action | System Response | Next State |
|-------|--------------|-----------------|------------|
| Browse | Scroll feed, tap Jump cards | Load chronological feed | Detail view |
| Judge | Tap "Judge", select 4 tiers | Validate eligibility (not own Jump, window open) | Judgment Receipt |
| Confirm | Review verdict, tap "Confirm and File" | Persist Judgment, update running average | Detail view (updated) |
| Create | Tap "Post Jump", capture Evidence | Open camera/photo picker | Compose |
| Compose | Enter Caption, Source, Destination, Food | Validate inputs, enable submit | Submit |
| Submit | Tap "Submit Jump" | Create Performed Jump, start 10-min Author Grace Period | Feed (Jump appears after grace period) |

### 1.4 Edge Cases in Core Loop

- **Author Grace Period (10 min):** After submission, the Jump appears on the performer's profile with an "Edit" badge. Other Players see "Judging Window opens in [countdown]". The performer may edit Caption or retract entirely.
- **Judging Window closed:** If a Jump's Season is finalized or the Jump is removed, the "Judge" button is replaced with "Judging Window Closed" (disabled state).
- **Self-judging:** The "Judge" button is hidden on the player's own Jumps. The detail view shows "You cannot Judge your own Jump" in the Judging panel area.
- **Already judged:** If the player has already Judged a Jump, the button reads "You have entered your Judgment" and is disabled. Judgments are final once confirmed and filed — there is no edit or re-Judge flow (ADR-0022).

## 2. Feed Model

### 2.1 Feed Architecture

The feed is **chronological, public, and algorithm-free**. Every Performed Jump appears in reverse chronological order. There is no follower graph, no recommendation engine, no "For You" tab.

**Design rationale:** The feed surfaces what exists; share velocity drives discovery. Algorithmic ranking would contradict the "Performance over Consumption" pillar by optimizing for time-on-app rather than Judging intent. A chronological feed also avoids the cold-start death spiral where an empty algorithmic feed kills engagement.

### 2.2 Jump Card Anatomy

Each Jump card on the feed displays:

```
┌─────────────────────────────────────┐
│ [Evidence Photo]                      │
│                                     │
│ Player Name    •  2h ago             │
│                                     │
│ Taco Bell  →  Olive Garden          │
│ Crunchwrap Supreme                   │
│                                     │
│ "Crunchwrap devoured in the Olive    │
│  Garden parking lot..."              │
│                                     │
│ 3.2  (12 Judgments)    [Judge →] │
└─────────────────────────────────────┘
```

**Elements per card (ordered by visual weight):**

| Element | Visual Weight | Tap Target | Navigates To |
|---------|--------------|------------|--------------|
| Evidence Photo | Primary — 16:9, cropped center, fills card width | Photo area | Jump Detail |
| Running Average Score + Judgment Count | High — right-aligned, large numeral | Score area | Jump Detail (score breakdown) |
| Source → Destination | Medium — icon + text, single line | Text area | Jump Detail |
| Food | Medium — below Source/Destination | Text area | Jump Detail |
| Caption | Low — truncated to 2 lines, "..." ellipsis | Text area | Jump Detail (full text) |
| Player Name + Timestamp | Lowest — lightweight attribution | Player name | Player Profile |
| Judge CTA | High — primary action button, right-aligned | Button | Judging Screen |

**Information hierarchy:** Photo > Score/Judgment count > Source/Destination/Food > Caption > Player name. The photo and score are the fastest-scan elements; the S/D/F line explains the Jump's premise at a glance.

**Card dimensions:**
- Evidence photo: full card width, 16:9 aspect ratio
- Card padding: 16px horizontal, 12px vertical between sections
- Card spacing: 8px between cards in feed
- Score numeral: 24pt bold; Judgment count: 14pt regular

### 2.3 Feed Navigation Patterns

- **Vertical scroll:** Standard infinite scroll, loading 20 Jumps per page
- **Pull-to-refresh:** Updates feed with latest Jumps; maintains scroll position
- **Tap card:** Opens Jump detail view (full-screen modal on mobile)
- **Swipe left on card:** Quick-share (opens native share sheet with deep link)
- **Empty state:** "No Jumps yet. Be the first to perform one." with "Post Jump" CTA

### 2.4 Discovery Without Algorithm

Discovery relies on three mechanisms:

1. **Chronological recency:** New Jumps appear at the top; active Players see fresh content on each return
2. **Share-driven deep links:** Shared Jumps open directly in detail view, bypassing the feed
3. **Weekly Prompt:** A curated Prompt appears as a pinned card at the top of the feed: "This week: Take breakfast food to a place of work." Tap opens Prompt detail with example Jumps and "Perform this Prompt" CTA

**No search, no hashtags, no categories in v1.** The feed is intentionally simple to maintain the "public stage" metaphor.

## 3. Onboarding Paths

### 3.1 Path A: Unauthenticated Visitor (Cold Open)

```
1. Open app → Feed Screen (no login gate)
   └─ Sees seeded Jumps, understands game from examples
   
2. Browse 2-3 Jumps → Tap "Judge" on any Jump
   └─ No auth required; Judging screen opens immediately
   
3. Submit Judgment → Judgment Receipt
   └─ "Your verdict has been entered into the record"
   
4. Return to Feed OR Tap "Post your own Jump"
   └─ If tap CTA → Auth Gate Screen (one-tap social login)
   
5. Auth Gate Screen
   └─ "Save your Judgments and compete in the Open"
   └─ [Sign in with Apple] [Sign in with Google] [Continue with Email]
   
6. Create Account → Guest Judgments migrate automatically
   └─ Return to Feed with "Post Jump" CTA enabled
```

**Visual design for "learn by example":** The feed IS the onboarding. No tutorial screens, no tooltips, no walkthrough. The Jump card design is self-explanatory: photo + Source/Destination/Food tells you what happened, score tells you how it was received, "Judge" button tells you what to do next. The first three Jump cards a visitor sees should be curated seed content that demonstrates the range of the game (one high-Transgression Jump, one high-Creativity Jump, one high-Commitment Jump) so the visitor infers the scoring dimensions from examples rather than instructions.

**Guest Judge identity cues:** After a Guest submits their first Judgment, a persistent banner appears at the top of the feed: "You've Judged 1 Jump on this device. Create an Account to save your history." The banner updates the count after each Judgment, making the accumulated investment salient. The banner is dismissible per session but reappears on next app open.

### 3.2 Path B: New Solo Player (Authenticated, No Group)

```
1. Open app → Auth Gate Screen (if not previously authenticated)
   └─ One-tap social login → Account created automatically
   
2. Display Name Setup → Set visible name → Continue

3. Feed Screen → Sees public feed + "Post Jump" FAB
   └─ Player has no Group; Groups are v2

4. Tap "Post Jump" → Create Jump Screen
   └─ Camera/photo picker → Caption → Source/Destination/Food

5. Submit Jump → Author Grace Period (10 min edit window)
   └─ Jump appears on feed after grace period expires

6. Receive Judgments → Push notification (if enabled): "Your Jump was Judged"
   └─ Tap notification → Jump Detail Screen (updated scores)

7. Browse feed, Judge others, share own Jump
   └─ Open eligibility: Player competes in monthly Open automatically
```

**Key decision:** Solo Players do not need a Group to participate. The Open provides competitive payoff without Group coordination. Groups are a v2 feature.

### 3.3 Path C: New Invited Player (Deep Link)

```
1. Receive share link in group chat → Tap deep link
   └─ Opens Jump Detail Screen directly (in app or web view)
   
2. Judge the shared Jump (as Guest or authenticated)
   └─ Same Judging flow as Path A

3. After Judging → Judgment Receipt
   └─ CTA: "Post your own Jump" (requires auth if Guest)
   └─ CTA: "Share Your Verdict" (shareable receipt card)

4. Return to Feed → Browse more Jumps
   └─ Standard Feed experience; no Group context in v1

5. If Player wants to post → Auth Gate (if Guest) → Create Jump
```

**Design rationale:** The invitation is Jump-first, not Group-first. The recipient experiences the core loop (Judge a Jump) before any social commitment. In v1, Groups are lightweight social circles with no formal infrastructure — there is no Group Home screen. The deep link delivers the recipient directly to the Jump, where they can Judge immediately. Group association is a v2 concept.

### 3.4 Onboarding Decision Map

| Entry Point | First Screen | First Action | Auth Gate Location | Conversion Goal |
|-------------|-------------|--------------|-------------------|-----------------|
| Cold open (unauth) | Feed | Browse/Judge | Posting or history claim | Guest-to-Player |
| Direct share link | Jump Detail | Judge | Posting or history claim | Share-to-Judge |
| App store/open | Feed | Browse/Judge | Posting or history claim | Guest-to-Player |

## 4. Screen Inventory (MVP)

### 4.1 Core Screens

| Screen | Purpose | Primary Action | Auth Required | Entry Points | Exit Points |
|--------|---------|---------------|---------------|-------------|-------------|
| **Feed** | Browse all public Jumps in reverse chronological order | Tap Jump card to view detail | No | App open; back from any detail screen; deep link fallback | Jump Detail (tap card); Judging (tap Judge on card); Create Jump (tap FAB); Share Sheet (tap Share) |
| **Jump Detail** | View full Jump: Evidence photo, Caption, Source/Destination/Food, running average, score breakdown, Judgment count | Judge / Share / Report | No (Judge), Yes (Share attribution) | Feed (tap card); Share deep link; Player Profile (tap Jump) | Feed (back); Judging Screen (tap Judge); Share Sheet (tap Share); Report Screen (tap Report); Player Profile (tap performer name) |
| **Judging** | Submit verdict on 4 factors via tap-to-select tier buttons | Select tier for all 4 factors, then confirm | No (Guest allowed) | Jump Detail (tap Judge) | Judgment Receipt (tap "Enter Judgment into Record"); Jump Detail (tap back/cancel) |
| **Judgment Receipt** | Review and confirm submitted verdict as a filing receipt | "Confirm and File" to submit Judgment | No | Judging Screen (tap "Enter Judgment into Record") | Jump Detail with updated scores (tap "Confirm and File") |
| **Create Jump** | Compose and submit a new Jump with Evidence, Caption, Source, Destination, Food | Capture/upload Evidence photo, fill fields, submit | Yes | Feed (tap FAB); Prompt Card (tap "Perform this Prompt") | Feed (after submit); Feed (tap cancel) |
| **Auth Gate** | Convert Guest to Player via one-tap social login | Select auth provider and authenticate | N/A (conversion screen) | Create Jump (if unauth); History claim; Guest cap prompt; Post-Judgment CTA | Create Jump (after auth); Feed (after auth); Feed (dismiss) |
| **Display Name Setup** | Set visible name (first-time only, after auth) | Tap "Continue" | Yes | Auth Gate (after successful auth) | Feed |
| **Player Profile** | View a Player's Jumps, stats, and Open Standing | Browse Player's Jump history | No (view), Yes (own profile edits) | Feed (tap player name); Jump Detail (tap performer name); Open Standings (tap player) | Jump Detail (tap Jump card); Feed (back) |
| **Open Standings** | Monthly competition rankings with score breakdowns | View rankings, tap Player profiles | No (view), Yes (compete) | Feed header icon; Feed (Open banner) | Player Profile (tap player row); Feed (back) |
| **Report** | Flag Jump for moderation with category selection | Select category, optionally add text, submit | Yes | Jump Detail (tap Report) | Jump Detail (after submit); Jump Detail (cancel) |
| **Tombstone** | Displayed when shared Jump has been removed | Tap "Browse Feed" | No | Deep link to removed Jump | Feed (tap CTA) |

### 4.2 Modal / Overlay Screens

| Screen | Purpose | Trigger | Dismissal |
|--------|---------|---------|-----------|
| **Share Sheet** | Distribute Jump via native share with deep link and preview card | Tap Share icon on Jump Detail | System share sheet dismissal |
| **Author Grace Period Banner** | Edit/retract own Jump within 10-minute window | Appears on own Jumps for 10 min post-submit | Auto-dismisses after grace period; tap to edit/retract |
| **Guest Cap Prompt** | Soft auth wall after 5 Guest Judgments | After 5th Guest Judgment | Auth Gate (tap CTA); Dismiss (tap X, reappears next session) |
| **Prompt Card** | Weekly curated Jump idea pinned to feed top | Always visible at top of Feed | Tap to expand; scroll past; auto-replaces weekly |
| **Post-Judgment CTA** | Convert engaged Guest Judge after submission | After Judgment Receipt for Guest Judges | Auth Gate (tap CTA); Dismiss (tap "Not now") |

### 4.3 Navigation Structure

```
Feed (root, default screen)
├── Jump Detail (push from Feed)
│   ├── Judging Screen (modal)
│   │   └── Judgment Receipt (modal)
│   ├── Report Screen (modal)
│   └── Player Profile (push)
├── Create Jump (modal, via FAB)
│   └── Auth Gate (modal, if unauth)
├── Open Standings (push from header icon)
│   └── Player Profile (push)
└── Player Profile (push from Feed)
    └── Jump Detail (push)
```

**Navigation rules:**
- **Maximum depth: 3 levels** from Feed root (e.g., Feed → Jump Detail → Judging → Judgment Receipt)
- **Modal for transient actions:** Judging, Judgment Receipt, Create Jump, Report, Auth Gate — these are tasks, not destinations
- **Push for exploration:** Jump Detail from Feed, Player Profile from Standings — these are content views
- **No tab bar in MVP.** The Feed is the primary (and nearly only) destination. A "+" FAB for composing and a header icon for Open Standings are the only persistent navigation elements. A tab bar implies multiple co-equal destinations that don't exist yet in MVP.
- **Back navigation:** Hardware back / swipe-back returns to previous level; modal close button (X) dismisses to parent

**Rationale for no tab bar:** MVP has one primary destination (Feed). Open Standings is accessible from a header icon. Profile is accessible by tapping a player name. Adding a 5-tab bar for screens that don't exist yet (Groups, Notifications) creates empty states and implies a more complex app than MVP delivers. A tab bar can be added in v2 when Groups ship.

## 5. Judging Interface

### 5.1 Layout

The Judging screen is a **single-screen, scrollable interface** showing all four factors simultaneously. No gestures, no sliders, no sequential per-factor screens (ADR-0022).

```
┌─────────────────────────────────────┐
│ ← Back                    [Help ?]  │
│                                     │
│ [Evidence Photo — 16:9 thumbnail]   │
│ Player Name • 2h ago                │
│ Taco Bell → Olive Garden            │
│ Crunchwrap Supreme                   │
│                                     │
│ ─────────────────────────────────── │
│ TRANSGRESSION                        │
│ How strongly does this Jump violate │
│ an expected food/place boundary?     │
│                                     │
│ ┌───┐ ┌───┐ ┌───┐ ┌───┐           │
│ │ 1 │ │ 2 │ │ 3 │ │ 4 │           │
│ └───┘ └───┘ └───┘ └───┘           │
│ Insufficient    Clear, intentional  │
│                                     │
│ ─────────────────────────────────── │
│ CREATIVITY                           │
│ How novel, thematic, or absurdly    │
│ elegant is this Jump?                │
│                                     │
│ ┌───┐ ┌───┐ ┌───┐ ┌───┐           │
│ │ 1 │ │ 2 │ │ 3 │ │ 4 │           │
│ └───┘ └───┘ └───┘ └───┘           │
│ Panel notes a   Structural          │
│ Jump occurred   connection revealed │
│                                     │
│ ─────────────────────────────────── │
│ COMMITMENT                           │
│ How completely did the performer    │
│ sell the bit with a straight face?   │
│                                     │
│ ┌───┐ ┌───┐ ┌───┐ ┌───┐           │
│ │ 1 │ │ 2 │ │ 3 │ │ 4 │           │
│ └───┘ └───┘ └───┘ └───┘           │
│ Performer       No indication of    │
│ found it        absurdity           │
│ funnier...                          │
│                                     │
│ ─────────────────────────────────── │
│ PRESENTATION                         │
│ How compellingly does the Evidence  │
│ capture the Jump as a performance?   │
│                                     │
│ ┌───┐ ┌───┐ ┌───┐ ┌───┐           │
│ │ 1 │ │ 2 │ │ 3 │ │ 4 │           │
│ └───┘ └───┘ └───┘ └───┘           │
│ Evidence is     Evidence is self-   │
│ present         sufficient          │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │  Enter Judgment into Record     │ │
│ └─────────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

**Layout rules:**
- **Factor order:** Transgression → Creativity → Commitment → Presentation. Transgression and Creativity appear above Presentation to avoid halo contamination from the visually anchored Presentation factor. Presentation is last because it is the most visually anchored factor and will create halo contamination on factors that follow it (ADR-0022).
- **Evidence thumbnail:** Compact 16:9 photo at top of screen (not full-width hero) to maximize space for factor sections. Player name, timestamp, and S/D/F summary appear below the thumbnail as context.
- **Factor section structure:** Each factor section contains: (1) factor name in ALL CAPS bold, (2) one-line description in regular weight, (3) four tier buttons in a horizontal row, (4) tier label text below the selected tier (or below the button row if space permits for all four labels).
- **Tier buttons:** Four large tappable buttons arranged horizontally. Selected state uses filled background + border + checkmark icon; unselected uses outline only. Minimum 44x44pt touch target per iOS HIG / 48x48dp per Material Design.
- **Tier labels:** Full label text displayed below the selected tier. Labels use the deadpan-institutional register (ADR-0022). When no tier is selected, the lowest and highest tier labels are shown as range indicators (e.g., "Insufficient ... Clear, intentional").
- **Scroll behavior:** Screen scrolls vertically if content exceeds viewport. Factor sections are visually separated by horizontal rules. The "Enter Judgment into Record" button is always visible at the bottom via sticky positioning when scrolling.
- **Sticky submit button:** The "Enter Judgment into Record" button pins to the bottom of the viewport. It is disabled (grayed out, no tap response) until all four factors have a selection. This prevents the user from having to scroll back down after completing the last factor.

### 5.2 Tap Targets and Interaction

- **Tier button size:** Minimum 44x44pt (iOS HIG) / 48x48dp (Material Design). Buttons expand equally to fill available width with 8px gaps.
- **Tap feedback:** Haptic light impact + visual fill animation (150ms ease-out). On reduced-motion, state change is instant with no animation.
- **Selection state:** Selected tier is filled (primary color background + white text + checkmark icon); unselected tiers are outlined (transparent background + border + dark text). Only one tier per factor may be selected. Tapping a different tier in the same factor replaces the selection.
- **Deselection:** Tapping the already-selected tier deselects it (returns to outline state). This allows the Judge to reconsider before submitting.
- **Validation:** All four factors must have a selection before "Enter Judgment into Record" is enabled. If the button is in disabled state, it shows a subtle hint: "Select a verdict for all four factors" below the button.
- **Progress indicator:** A subtle progress bar or step indicator at the top of the screen shows how many of 4 factors have been selected (e.g., "2 of 4 factors evaluated"). This reduces the cognitive load of tracking completion.

### 5.3 Confirmation Flow

After tapping "Enter Judgment into Record":

```
┌─────────────────────────────────────┐
│ ← Back                              │
│                                     │
│ JUDGMENT RECEIPT                     │
│                                     │
│ Jump: Taco Bell → Olive Garden      │
│ Crunchwrap Supreme                   │
│                                     │
│ Your verdict:                        │
│                                     │
│ Transgression    4                   │
│ Violation is clear, intentional,    │
│ and defensible under no             │
│ conventional logic                   │
│                                     │
│ Creativity       3                   │
│ The connection is visible; a        │
│ reasonable person would not have    │
│ made it by accident                  │
│                                     │
│ Commitment       3                   │
│ Commitment holds; a single lapse    │
│ does not materially undermine the   │
│ record                               │
│                                     │
│ Presentation     2                   │
│ Evidence is present; the Caption    │
│ carries significant documentary     │
│ weight                               │
│                                     │
│ [Confirm and File]                    │
│                                     │
└─────────────────────────────────────┘
```

**Confirmation screen rules:**
- **Title:** "Judgment Receipt" (deadpan-institutional register)
- **Content:** Full Jump summary (Source → Destination, Food) + all four verdicts with tier number and full label text
- **Layout:** Each factor shows tier number right-aligned, full label text below. Factors are separated by subtle dividers.
- **Actions:** "Confirm and File" (primary, full-width button). No "Edit" action — Judgments are final once filed (ADR-0022).
- **Post-submit:** Brief toast "Judgment entered into the record" (2s), then return to Jump Detail with updated running average
- **No celebration:** No confetti, no animations, no "Great job!" The tone is institutional and serious, reinforcing that Judging is a performance, not a game reaction (Product Vision, "Judgment as Play" pillar).

### 5.4 Error States

| Error | Trigger | UI Treatment | Recovery |
|-------|---------|-------------|----------|
| **No network** | Submit Judgment while offline | Inline banner: "Unable to file Judgment. Check connection and try again." | Retry button; auto-retry on reconnect |
| **Already judged** | Player tries to Judge a Jump they already Judged | "Enter Judgment" button reads "You have entered your Judgment" and is disabled | None; Judgments are final once filed |
| **Judging Window closed** | Season finalized or Jump removed | Disabled "Judge" button with text: "Judging Window Closed" | None; inform user |
| **Self-judging** | Tap "Judge" on own Jump | Hidden button; inline text: "You cannot Judge your own Jump" | None |
| **Grace period active** | Tap "Judge" within 10 min of submission | Button disabled; countdown: "Judging Window opens in [MM:SS]" | Wait for countdown |
| **Guest cap reached** | 5th Guest Judgment | Modal: "You've Judged 5 Jumps. Create an Account to save your history and compete in the Open." | Auth Gate or dismiss |

### 5.5 Accessibility Considerations

- **Screen reader support:** Each tier button has `accessibilityLabel` describing factor + tier value + label text (e.g., "Transgression, tier 4, Violation is clear, intentional, and defensible under no conventional logic")
- **Focus management:** On Judging screen open, focus moves to first unselected factor. After selecting a tier, focus moves to next unselected factor. All four selected → focus moves to "Enter Judgment into Record" button.
- **Dynamic type:** All text scales with system font size. Tier buttons expand vertically to accommodate larger text; horizontal layout becomes 2x2 grid if needed.
- **Color independence:** Selected state uses fill + border + icon (checkmark), not color alone. Unselected uses outline only.
- **Reduced motion:** Disable haptic and fill animation when reduced motion is enabled. Selection state changes instantly.
- **VoiceOver/ TalkBack:** Full label text is read for each tier. Factor descriptions are read as hints, not labels, to avoid verbosity.

## 6. Share / Deep-Link UX

### 6.1 Share Card Specification

When a Player shares a Jump, the native share sheet surfaces a preview card. The card must be intelligible to someone who has never heard of Supperjumpin.

**In-app share card (rendered by the app for the share sheet):**

```
┌─────────────────────────────────────┐
│ [Evidence Photo — 1.91:1 OG ratio]  │
│                                     │
│ "Crunchwrap devoured in the Olive    │
│  Garden parking lot..."              │
│                                     │
│ 3.2  (12 Judgments)               │
│                                     │
│ Taco Bell  →  Olive Garden          │
│ Crunchwrap Supreme                   │
│                                     │
│ Tap to Judge → supperjumpin.app/j/abc│
└─────────────────────────────────────┘
```

**Card elements (ordered top to bottom):**

| Element | Specification | Purpose |
|---------|--------------|---------|
| Evidence Photo | 1.91:1 aspect ratio (Open Graph standard), cropped center | Visual hook — the absurd image is the share's primary value |
| Truncated Caption | Max 2 lines, 140 characters, "..." ellipsis | Context — explains what's happening without requiring the full detail |
| Running Average Score | Large numeral (e.g., "3.2") + "(12 Judgments)" in smaller text | Social proof — shows the Jump has been evaluated |
| Source → Destination | Icon + text, single line | Premise — immediately communicates the Jump's concept |
| Food | Text below Source/Destination | Specificity — names the food being moved |
| Deep Link URL | `supperjumpin.app/j/{jump_id}` | Action — one tap to view and Judge |
| Viral nudge line | "Only 3 people have judged this Jump. Add your score." (conditional) | Friction reduction — lowers the barrier to first Judgment |

**Platform-specific adaptations:**

| Platform | Image Ratio | Text Limit | Link Format | Special |
|----------|------------|------------|-------------|---------|
| iMessage | 1.91:1 (OG) | 2 lines Caption + S/D/F | Deep link | Rich link preview with app badge |
| Twitter/X | 16:9 or 1.91:1 | 280 chars total (Caption + S/D/F + URL) | `t.co` wrapped deep link | Card shows photo + first line of Caption |
| Instagram Stories | 9:16 (story) | Overlay text: Score + S/D/F | "Link in bio" or swipe-up | App generates story-ready image with score overlay |
| General web (OG meta) | 1.91:1 | `og:title` = S/D/F summary; `og:description` = truncated Caption | Deep link URL | Standard Open Graph meta tags |

**Share card aspect ratio decision:** 1.91:1 is the primary ratio because it is the Open Graph standard and renders correctly across iMessage, Twitter, Slack, Discord, and most link preview systems. Instagram Stories require a separate 9:16 image, generated on-device at share time.

### 6.2 Recipient Experience

**Scenario A: Recipient has app installed**
- Tap deep link → App opens directly to Jump Detail Screen
- Can Judge immediately (Guest or authenticated)
- CTA to "Post your own Jump" appears after Judging

**Scenario B: Recipient does not have app installed**
- Tap deep link → Mobile web landing page showing Jump Detail
- "Judge this Jump" button opens lightweight web Judging interface (same 4-factor tap-to-select)
- After Judging, prompt: "Download the app to save your Judgment and see more"
- App store link + "Continue as Guest" option (Judgment stored by session)

**Scenario C: Link opened in desktop browser**
- Web view of Jump Detail with Evidence photo, Caption, scores
- "Scan QR code to Judge on your phone" (QR code deep link)
- No desktop Judging in v1

### 6.3 Fallbacks

| Condition | Fallback Behavior |
|-----------|------------------|
| Deep link expired | Show feed with toast: "This Jump is no longer available" |
| Jump removed | Tombstone page: "This Jump has been removed" (no content, no performer info) |
| Network error | Cached share card if available; otherwise generic "Unable to load Jump" |
| Guest Judge on web | Lightweight web Judging; prompt to download app after submission |
| App not installed | Web landing page → app store prompt; preserve deep link intent for post-install |

### 6.4 Attribution

Every share link includes:
- `sharer_player_id`: Who shared it
- `jump_id`: Which Jump
- `channel`: How it was shared (iMessage, Instagram, etc.) — detected where possible

This enables tracking of Share-to-Judge rate and viral coefficient (K).

### 6.5 Share Card as Onboarding Artifact

The share card is the primary onboarding surface for invited users. It must communicate three things in under 3 seconds:

1. **What happened:** Photo + Source/Destination/Food tells the story
2. **That people care:** Score + Judgment count provides social proof
3. **What to do:** "Tap to Judge" is the clearest possible CTA

The share card does NOT explain the scoring factors, the four-tier system, or the game's rules. It relies on the photo's inherent absurdity to create curiosity, and the "Judge" CTA to pull the recipient into the app where the full Judging interface teaches the system through action.

## 7. Usability Risks and Mitigations

### 7.1 Risk: Judging Feels Like Work

**Evidence:** BeReal's daily notification became "homework" for some users. Supperjumpin's four-factor Judging is more cognitively demanding than a Like or a swipe.

**Mitigation:**
- Single-screen layout keeps Judging under 30 seconds
- Tap-to-select is faster than sliders or gestures
- No mandatory Judging — Players can browse indefinitely without Judging
- "Judge Back" prompt (MVP Roadmap) creates reciprocity without coercion: "3 people Judged your Jump. Return the favor?"

### 7.2 Risk: First Jump Anxiety

**Evidence:** Research shows "fear of posting" is the #1 barrier for lurkers. Performing a Jump requires real-world effort and social exposure.

**Mitigation:**
- Weekly Prompt reduces blank-page syndrome: "This week: Take breakfast food to a place of work"
- First Jump is not labeled "practice" (that would devalue the public stage), but the Author Grace Period allows retraction within 10 minutes
- Guest Judges can experience the full loop before committing to auth
- "Jump Kits" (Post-MVP) will suggest specific combinations

### 7.3 Risk: Empty Feed at Cold Start

**Evidence:** Pre-moderation creates a cold-start death spiral. An empty feed kills engagement.

**Mitigation:**
- Seed 20–30 Jumps from founders/friends before inviting Players
- No pre-moderation — Jumps appear immediately after Author Grace Period
- Weekly Prompt ensures fresh content rhythm even with few Players

### 7.4 Risk: Transgression Escalation

**Evidence:** The Transgression scoring axis structurally rewards pushing boundaries. TikTok challenge history shows this pattern has real-world consequences.

**Mitigation:**
- House Rules are linked from the Jump Composer (info icon) and the Report screen
- "Be mindful" nudge at submission time for high-Transgression Jumps (Post-MVP safety experiment)
- Report button on every Jump (4 categories + Other)
- Manual team moderation at MVP scale; auto-hide on multiple reports deferred to Post-MVP

### 7.5 Risk: Guest Judges Feel Invisible

**Evidence:** Guests have no social identity in the app. They are invisible to other Players and to themselves across sessions.

**Mitigation:**
- Guest attribution: Jump Authors see "Guest" as a Judge (not anonymous — just unauthenticated)
- Persistent Guest banner: "You've Judged 3 Jumps on this device. Create an Account to save them."
- Judgment receipt share: Guests can share their verdict as a card, creating identity investment
- Soft Guest cap (5 Judgments, server-configurable) creates a natural decision point without being punitive

### 7.6 Risk: The Open Feels Pointless at Small Scale

**Evidence:** Research shows global leaderboards without social context demotivate 99% of users. At 100–500 Players, a monthly global Standing is structurally weak.

**Mitigation:**
- Treat the Open as a signal surface, not a retention engine (MVP Roadmap)
- Instrument aggressively: Standing check rates, Judgment velocity, posting cadence
- Be prepared to pivot in Month 2–3: weekly mini-Opens, cohort segmentation, or Prompt challenges
- Do not rely on the Open for retention — invest in push notifications, share artifacts, and content quality

## 8. Visual Design System Decisions

This section codifies the visual design decisions that are load-bearing for the product experience. Decisions that are purely aesthetic or implementation-detail are deferred to a separate design system specification.

### 8.1 Decisions Codified in This Document

These decisions directly affect the product experience and must be consistent across all screens:

**Tone register (visual):** The app's visual language mirrors the deadpan-institutional tone of the game's terminology. Clean, structured layouts with generous whitespace. No playful illustrations, no cartoon icons, no confetti. The humor comes from the content (absurd food photos), not the chrome. The app should feel like a well-run bureaucratic institution that happens to adjudicate Crunchwrap placement.

**Color palette (semantic):**
- **Primary (institutional):** Dark brown `#2f241d` — headings, primary buttons, borders. Evokes paper, wood, institutional gravitas.
- **Secondary (warm):** Terracotta `#c1673a` — accents, secondary actions, input borders. Warm without being playful.
- **Background (parchment):** Cream `#f7efe2` — screen backgrounds. Reads as paper, not white.
- **Surface (card):** Off-white `#fffaf2` — card backgrounds, elevated surfaces.
- **Success (recorded):** Muted green `#2f7d2f` — Judgment confirmation, recorded states. Institutional green, not celebration green.
- **Text primary:** Dark brown `#2f241d` — body text, headings.
- **Text secondary:** Medium brown `#4d3b31` — captions, descriptions, secondary text.
- **Error:** Deep red — validation errors, report categories. Used with icon, never color alone.

These colors are derived from the existing prototype (App.tsx StyleSheet) and codified here as the MVP palette. A full design system document may expand the palette with additional semantic tokens (e.g., `--color-factor-transgression`, `--color-factor-creativity`) but the base palette is locked.

**Typography (scale):**
- **Display:** 36pt, weight 900 — app title only
- **Heading:** 24pt, weight 900 — screen titles
- **Section:** 18pt, weight 800 — section headers, factor names
- **Body:** 16pt, weight 400 — body text, descriptions
- **Caption:** 14pt, weight 400 — timestamps, Judgment counts, secondary info
- **Score:** 24pt, weight 900 — running average numerals

Font family is deferred to the design system spec. The scale is locked; the family is not.

**Spacing (grid):** 4px base unit. All spacing is a multiple of 4px.
- 4px: inline gaps
- 8px: between related elements (tier buttons, list items)
- 12px: card internal padding
- 16px: card external padding, section gaps
- 24px: screen padding, major section separation

**Border radius:**
- 24px: cards (primary surface)
- 12px: input fields, Jump cards, buttons
- 8px: tier buttons, small interactive elements

**Evidence photo treatment:**
- Aspect ratio: 16:9 in feed and Jump Detail; 1.91:1 in share cards
- Crop: center crop (no letterboxing)
- No filters, no overlays in v1 — the photo speaks for itself

### 8.2 Decisions Deferred to Design System Spec

These are implementation details that do not affect the product experience at the decision level:

- Font family selection (system font vs. custom)
- Exact animation curves and durations (beyond the 150ms tier-fill specified above)
- Icon set selection
- Dark mode palette
- Component-level API design (props, variants, composition patterns)
- Gluestack UI v2 theme configuration (ADR-0014)
- Shadow and elevation system
- Micro-interaction library (button press states, loading skeletons)

### 8.3 Component Structure (Screen-Level)

Each screen decomposes into reusable components. This is the component ownership map, not the full component API:

| Screen | Components |
|--------|-----------|
| Feed | `FeedList`, `JumpCard`, `PromptCard`, `GuestBanner`, `PostJumpFAB` |
| Jump Detail | `EvidenceHero`, `JumpMeta` (Source/Destination/Food), `CaptionBlock`, `ScoreDisplay`, `JudgeButton`, `ShareButton`, `ReportButton` |
| Judging | `JudgingHeader` (thumbnail + meta), `FactorSection` ×4, `TierButton` ×4 per factor, `SubmitJudgmentButton` |
| Judgment Receipt | `JudgmentReceipt`, `VerdictRow` ×4, `ConfirmButton` |
| Create Jump | `PhotoCapture`, `CaptionInput`, `LocationInput` (Source, Destination), `FoodInput`, `SubmitJumpButton` |
| Auth Gate | `AuthPrompt`, `SocialLoginButton` (Apple, Google), `EmailMagicLink` |
| Display Name Setup | `DisplayNameInput`, `ContinueButton` |
| Player Profile | `ProfileHeader`, `JumpList`, `StatBlock` |
| Open Standings | `StandingsHeader` (month, countdown), `StandingRow`, `PlayerLink` |
| Report | `ReportCategoryList`, `ReportTextField`, `SubmitReportButton` |
| Tombstone | `TombstoneMessage`, `BrowseFeedButton` |

**Shared components used across screens:**
- `JumpCard` — used in Feed, Player Profile
- `ScoreDisplay` — used in Jump Card, Jump Detail, Share Card
- `PlayerNameLink` — used in Jump Card, Jump Detail, Standings
- `SourceDestinationFood` — used in Jump Card, Jump Detail, Share Card, Judgment Receipt

## 9. Accessibility

### 9.1 Judging Interface Accessibility

The Judging interface is the most critical accessibility surface because it is the core loop and must work for all users.

**Screen reader flow:**
1. Open Judging screen → "Judging screen. Four factors to evaluate. Swipe right to begin."
2. Focus on first factor → "Transgression. How strongly does this Jump violate an expected food place boundary? Four tiers available."
3. Swipe to tier 1 → "Tier 1. The panel finds insufficient transgression to warrant elevated consideration. Double tap to select."
4. Select tier → "Selected. Transgression, tier 1. Swipe right for next factor."
5. Repeat for all factors → "All factors evaluated. Enter Judgment into Record button."

**Keyboard navigation (external keyboard / switch control):**
- Tab moves between tier buttons and primary action
- Space/Enter selects tier or submits
- Escape cancels and returns to Jump Detail

### 9.2 Feed Accessibility

- Jump cards have `accessibilityLabel` summarizing: "[Player Name] performed a Jump. [Food] from [Source] to [Destination]. Score [X] based on [N] Judgments."
- Double-tap card opens detail view
- Three-finger swipe scrolls feed (standard iOS behavior)

### 9.3 Color and Contrast

- All text meets WCAG 2.1 AA contrast ratios (4.5:1 for normal text, 3:1 for large text)
- Selected tier buttons use fill + border + checkmark icon (not color alone)
- Error states use icon + text (not red alone)

### 9.4 Motion and Animation

- All animations respect `prefers-reduced-motion`
- Judging tier selection: instant state change when reduced motion is on; 150ms fill animation otherwise
- No auto-playing video or parallax effects

## 10. References and Contradictions

### 10.1 References

This document directly references and depends on:

- **Product Vision (01-product-vision.md):** Design pillars, target early Player, first 5-minute experience, growth model
- **ADR-0019 (Jumps Are Public By Default):** Public feed, no Group requirement for viewing or Judging
- **ADR-0022 (Judgment Interaction Model):** Single-screen tap-to-select, 1–4 named tiers, confirmation receipt, factor order
- **ADR-0023 (The Open):** Monthly competition, soft-close, platform-run
- **ADR-0024 (House Rules):** Safety boundaries, Report flow, Removed Jump behavior
- **ADR-0014 (Gluestack UI):** Component library for Expo app
- **ADR-0006 (Supabase Auth):** Auth providers, magic links, social login
- **MVP Roadmap (04-mvp-roadmap.md):** Scope boundaries, success metrics, deferred features
- **Guest Judge Conversion Analysis (docs/research/guest-judge-conversion-analysis.md):** Auth model contradiction, conversion mechanics, benchmarks
- **Open Competitiveness Analysis (docs/analysis/open-competitiveness-analysis.md):** Leaderboard design risks, signal-first approach
- **Judge-to-Performer Transition Analysis (docs/analysis/judge-to-performer-transition.md):** Micro-contribution ladder, bridge mechanics
- **Growth Loop Analysis (docs/design/growth-loop-analysis.md):** Share artifact design, viral nudge infrastructure

### 10.2 Flagged Contradictions

| # | Contradiction | Source A | Source B | Resolution in This Document |
|---|--------------|----------|----------|----------------------------|
| 1 | **Auth gates Judging vs. Guest Judges allowed** | Product Vision (01): "hits a soft auth gate when they try to Judge" | CONTEXT.md + MVP Roadmap: "In v1, all Judging is available to Guest Judges" | **Resolved:** This document follows the MVP Roadmap. The first auth wall is at posting/history claim (Step 5), not Judging (Step 3). A soft Guest cap (5 Judgments, server-configurable) creates a later decision point. The Product Vision's description is aspirational for v2. |
| 2 | **Judging interface: gestures vs. tap-to-select** | Early prototype (demo-script.md): "Gesture scoring, PanResponder swipes" | ADR-0022: "No gestures, no sliders, no sequential per-factor screens" | **Resolved:** This document implements ADR-0022 exclusively. The gesture model is superseded. |
| 3 | **Scoring factors: Difficulty vs. Commitment** | Early docs/demo-script: "Difficulty, Transgression, Creativity, Documentation" | ADR-0022: "Commitment, Transgression, Creativity, Presentation" | **Resolved:** This document uses the ADR-0022 factor set. Documentation was renamed to Presentation in ADR-0020; Difficulty was replaced by Commitment. |
| 4 | **Group-first vs. public-first** | Early demo-script: "Player creates or joins a Group" as step 2 | ADR-0019: "Jumps are public by default; Groups are optional overlay" | **Resolved:** This document follows ADR-0019. There is no Group tab or Group Home in MVP. Solo Players compete in the Open without Group membership. Groups are v2. |
| 5 | **The Open as retention engine vs. signal surface** | Product Vision: "The Open solves the payoff gap... gives Players a reason to return" | Open Competitiveness Analysis: "The Open alone does not fill this gap... not compelling enough to drive retention" | **Resolved:** This document treats the Open as a signal surface with ritual value, not a primary retention driver. Retention is carried by push notifications, share artifacts, and content quality. The Open's true value is behavioral signal. |

## 11. Open Questions

These questions are raised by this design but cannot be resolved without implementation and user data:

1. **What is the optimal Guest Judgment cap?** This document specifies 5 as the v1 default (server-configurable). The exact number should be A/B tested against Guest-to-Player conversion rate. If drop-off at the cap is too high, raise to 7; if conversion is too low, lower to 3.
2. **Does the weekly Prompt belong as a pinned card or a separate tab?** A pinned card preserves feed simplicity; a tab increases discoverability but adds navigation complexity.
3. **How should the feed handle the Author Grace Period?** Should the Jump be visible immediately with an "Editing" badge, or hidden entirely for 10 minutes? This document recommends visible with badge to maintain chronological flow.
4. **What is the optimal share card aspect ratio?** 1.91:1 (Open Graph standard) is recommended here for cross-platform compatibility, but Instagram Stories requires 9:16. Should the app generate both at share time, or prioritize one?
5. **Should Guest Judges see their own Judgment history without auth?** This document assumes no — history claim requires auth. But a local-only history (device-scoped) could increase conversion by making the loss salient.
6. **Should tier labels show for all four tiers simultaneously, or only for the selected tier?** Showing all four labels provides more context but increases vertical space per factor. This affects whether the Judging screen requires scrolling on smaller devices.
7. **What is the right visual treatment for the "Judging Window opens in [countdown]" state?** A countdown timer on the Judge button is clear but may create anxiety. A simpler "Available soon" message is calmer but less informative.

---

*Document status: Complete. Parent tracker: #106. Depends on: Product Vision (01), ADR-0019, ADR-0022, ADR-0023, ADR-0024, MVP Roadmap (04).*
