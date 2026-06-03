# Brand direction prototype — issue #185

**Status:** throwaway prototype, awaiting verdict. Not production. Not React Native.
**Run:** `cd apps/mobile/prototype-brand-185 && python3 -m http.server 8185` → http://localhost:8185

## The question

> Does **Modern 90s Food Court Arcade** land as the Jump-detail brand, and what
> do we steal from the supporting directions?

Issue #185 already stack-ranked the directions — this prototype is **not** re-picking
the brand. Food Court Arcade is the decided primary. The job is to see it executed at
fidelity, sanity-check it against real screen density, and surface the
"combined direction candidate" the issue itself proposed.

## Why web, not React Native (the medium decision)

The real app is Expo RN, and `prototype/UI.md` prefers hosting variants in the real
app. We deliberately broke that rule here: the effects this brief is *made of* —
Bungee/Anton/Shrikhand display type, ticket-perforation edges, grain/laminate
overlays, disposable-camera borders, the "score chips drop into four slots → final
score pops" motion — are trivial in HTML/CSS and genuinely painful in RN StyleSheet.
Forcing them through RN would *understate* each direction and defeat a brand
exploration. So: standalone HTML, vanilla-JS `?variant=` switcher, one phone frame.

To avoid the "everything looks fine in a vacuum" trap that UI.md warns about, every
variant renders the **same real Jump-detail payload** (mirrors the `JumpDetail` schema
in `packages/api-client` and the fields `JumpDetailScreen.tsx` reads) and is testable
against the **real `viewerContext` states** via the state bar.

## The five variants

| Key | Direction | Role per #185 | Structural identity |
|-----|-----------|---------------|---------------------|
| **A** | **Food Court Arcade** | **Primary brand (deepest build)** | Tall combo-menu / prize-ticket card; banner + perforation + score chips |
| B | Bowling / Arcade League | Supporting → **Standings** | Dark high-score *table*; numbers lead, photo secondary |
| C | Disposable Camera Field Report | Supporting → **Evidence** | Photo-forward manila case file; typewriter + handwritten caption |
| D | Extreme Sports, But Dumb | Supporting → **badges / share cards** | Full-bleed cinematic hero; giant overlaid stat |
| E | 50s Diner Dare League | Demoted alternate | Printed receipt/check; line-item totals, FILED stamp |

Switch: floating bottom bar, `←`/`→` keys, or `?variant=A..E`.

### Testing against real density (do this when reviewing)
State bar (bottom-right) flips the **same** design through the real judging states —
this is the honest test, not the hero card:
- **Can judge** → four score-chip slots + LOCK IT IN
- **Grace** → author-grace countdown (live ticker), judging not yet open
- **Judged** → disabled affordance + "score added to the board" / LOCKED IN stamp
- **Own jump** → "you can't judge your own Jump"
- **Disputes** / **Final score** toggles (independent)

Deep-link any state: `?variant=A&state=grace&disputes=1&final=1&scroll=560`

**Verification done:** A spot-checked across all viewer states (can-judge / grace /
judged+disputes+final) with screenshots — the aesthetic holds under each. B–E are
code-complete with the same state branches but were visually verified only in the
default can-judge state.

**Reviewer tip:** the state bar is fixed chrome. On a narrow window it docks to the
top as a pill; widen the browser past ~1100px and it moves to the right gutter beside
the device. Either way it should never cover the judging CTA.

**Label hygiene:** the issue's suggested labels were brainstormed *before* checking
CONTEXT.md's avoid-lists — e.g. "CHALLENGE FILED" uses `challenge`, an avoided Jump
synonym (swapped to "JUMP FILED" here). **Run every #185 label through the CONTEXT.md
avoid-lists before any of them enter the brand doc.**

## Recommended verdict to confirm (the combined-direction candidate)

Matches the issue's own proposal, now that it's visible at fidelity:

- **A (Food Court Arcade) = the house style.** Banner + Bungee jump number, combo-menu
  "THE ORDER", the four-slot score-chip judging mechanic, prize-ticket score counter,
  LOCKED IN stamp. This is the Jump-detail target screen the issue asked for.
- **Steal B's high-score table for Standings/Open leaderboards** — A's card language
  doesn't scale to a ranked list; B's row-based scoreboard does.
- **Steal C's disposable-camera frame + timestamp burn-in for the Evidence surface**
  (capture/review screens, share cards). A already borrows a light version of it.
- **D's giant-stat hero is the share-card / "final score reveal" treatment**, not the
  default detail screen (too photo-hungry and dark for everyday density).
- **E (diner) is out** as the main lane — confirmed; keep only the rubber-stamp
  "FILED" confirmation gesture, which all directions already share.

**The actual decision needed from a human:** confirm A as primary + that the steal-list
above is the v1 design-system seed. Then turn it into the lightweight brand doc the
issue's acceptance criteria call for.

## ⚠️ RN port cost (hidden implementation tax — read before adopting)

This web prototype can sell effects that are real work in the Expo RN screen. Budget for:

- **Display fonts** (Bungee / Anton / Shrikhand / Lilita / Special Elite / Caveat):
  need `expo-font` + bundled `.ttf`s. Not free like a Google Fonts `<link>`.
- **Ticket / receipt perforation edges** (A's dashed perf-edge, E's CSS `mask`): no CSS
  `mask`/`clip-path` in RN. Use `react-native-svg` (mask or a notched `<Path>`) or
  pre-baked PNG/9-patch frames. The notched prize-ticket is the most-reused motif —
  worth one solid SVG component.
- **Grain / laminated-tray texture** (A's `::before` overlay): ship a tiling PNG with
  low opacity; no `mix-blend-mode` in RN core.
- **Disposable-camera timestamp burn-in** (C/A): an absolutely-positioned monospace
  `<Text>` over the `<Image>` — cheap, port as-is.
- **Score-chip drop / count-up / slam-in / LOCKED IN stamp**: `react-native-reanimated`
  (spring + sequence). The signature judging animation is the one worth doing properly;
  the rest can ship static first.
- **`box-shadow` offsets** (A's hard purple drop-shadow, the chunky button shadows):
  RN shadows are softer/uglier; approximate with a stacked offset `<View>` or skip.

Net: A is adoptable, but ~the perforation SVG + the reanimated judging sequence + the
font bundling are the three line items that need real engineering, not a CSS port.

## Cleanup

When the verdict is captured into a brand doc / theme tokens, **delete this whole
`prototype-brand-185/` folder** (and stop the http.server). Don't promote this code —
it was written under prototype constraints (no tests, hotlinked Unsplash image,
inline styles). Rewrite properly when folding into the RN app.

## Side note (out of scope here)
CLAUDE.md's transitional-state table still says mobile is "single-file App.tsx", but it's
now App.tsx + FeedScreen.tsx + JumpDetailScreen.tsx. Worth a one-line fix in a separate PR.
