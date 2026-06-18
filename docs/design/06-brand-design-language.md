# Brand Design Language

Parent tracker: #185

## Decision

Supperjumpin should use a **Modern 90s Food Court Arcade** design language: a crisp, readable mobile game UI with food-forward hero imagery and fast-food arcade energy.

The working brand sentence is:

> A food-court arcade game for laughing at your friend's ridiculous bit.

This replaces the earlier faux-official emphasis. Supperjumpin should not feel like a court, case file, receipt, council, or bureaucracy. The structure should make the game easy to understand; the tone should stay fun, low-stakes, and social.

## Visual Foundation

Use the A/D prototype feedback as the seed:

- **A owns the system foundation**: colors, typography, readability, hierarchy, and day-to-day interface clarity.
- **D owns the food-forward energy**: large food imagery, memorable hero moments, and shareable visual punch.
- The final direction is a hybrid: **A's clarity with D's food-as-centerpiece composition**.

Food imagery should be treated as the most important artifact inside the UI, not as an uncontrolled background for text.

## Metaphor Stack

### Primary Structure: Arcade Cabinet / Challenge Board

Use this for hierarchy, action framing, status panels, score framing, and satisfying controls.

Good motifs:

- cabinet-like frames
- challenge board panels
- status lights
- chunky arcade controls
- score windows
- quick button feedback
- short result panels

### Content Organization: Fast-Food Combo Menu / Order Board

Use this for arranging Jump content clearly.

Good motifs:

- combo-menu sections
- overhead-board alignment
- bold item titles
- compact metadata chips
- clear grouped rows for food, source, destination, caption, score, and call state

### Accent Layer: Prize Ticket / Arcade Reward

Use this sparingly for score/reward moments, badges, confirmations, and shareable celebration.

Prize-ticket styling is seasoning, not the plate.

## Loudest Surfaces

The brand should be loudest on:

1. **Jump Detail / Jump Card**
2. **Make the Call flow**

These are the primary proof-of-language screens. If the design works there, Feed, Evidence capture, and Standings can inherit the system more calmly.

## Jump Detail / Jump Card

Jump Detail should use one canonical card system whose emphasis changes by state:

- Active states emphasize the next action, food image, caption, and call status.
- Completed states emphasize final score, outcome, food image, and shareability.

Layout direction:

- Large food image first, framed by a crisp arcade/challenge-board shell.
- Functional text never sits directly on the photo.
- Jump title, player, source, destination, food, caption, score, and state live in high-contrast panels or chips.
- The photo supplies the chaos; the UI supplies the order.

## Make the Call Flow

The judging interaction is the retention loop and should feel immaculate.

User-facing direction:

> React to your friend's ridiculous bit, then lock in that reaction with a satisfying arcade move.

Use **Make the Call** as the current user-facing action language in design examples. Keep `Judgment` as the internal/domain term until #308 decides whether deeper language changes are warranted.

Good CTA and confirmation examples:

- Make the Call
- Call It
- Lock It In
- Call locked
- Your call is on the board
- Nice. Next Jump?

Avoid making the user feel like a referee, panel judge, council member, or official authority. The user is primarily an audience member laughing with the group, secondarily an arcade player pressing a satisfying button.

When judging clarity conflicts with visual energy, sacrifice in this order:

1. decorative chaos
2. color saturation
3. display typography
4. image size
5. motion

Motion is worth preserving when it is fast, tactile, and reinforces completion. It must never slow down the task.

## Readability Rules

These are non-negotiable:

- Functional text never goes directly on photos in core product UI.
- Titles, criteria, scores, CTAs, labels, and instructions must sit on solid surfaces, panels, chips, plates, or standardized scrims.
- Large decorative labels may overlap imagery only with approved contrast protection.
- Body and CTA text should meet WCAG AA contrast.
- Images may create energy; surfaces must preserve comprehension.

## Initial Tokens

Start with the issue #185 palette, but treat the colors as semantic roles, not decoration.

```css
--sj-taco-purple: #5B2A86;
--sj-baja-teal: #00A6A6;
--sj-pizza-red: #E13A2D;
--sj-nacho-yellow: #FFC928;
--sj-hot-pink: #F15BB5;
--sj-deep-ink: #211A2E;
--sj-tray-white: #FFF7E8;
--sj-cup-blue: #2D9CDB;
--sj-lettuce-green: #7AC943;
```

Suggested roles:

- `tray-white`: primary app and panel background
- `deep-ink`: primary text
- `pizza-red` or `baja-teal`: primary action
- `taco-purple`: frame, shadow, or high-emphasis brand structure
- `nacho-yellow` and `hot-pink`: celebration accents, not large reading surfaces
- `cup-blue` and `lettuce-green`: secondary accents or status support

## Typography

Use a two-tier type system:

- Display font for short brand moments, Jump titles, headers, badges, and score moments.
- UI font for anything the user must read quickly, trust, or act on.

Recommended starting point from #185:

- Headlines/display: Bungee
- Body/UI: Nunito Sans
- Scores/meta: IBM Plex Mono

Rule: if the user must act on it, use the UI font.

## Motif Whitelist

Preferred motifs:

- arcade cabinet / challenge board structure
- fast-food combo menu / order board organization
- food image as hero artifact
- big reaction buttons / arcade controls
- score windows and status lights
- prize tickets as score/reward accents only
- chunky but readable display type
- warm tray-white surfaces
- crisp high-contrast panels

## Motif Blacklist

Avoid these as primary metaphors or visible copy directions:

- case files, evidence folders, investigation boards, courtrooms, councils, rulings
- receipts, order slips, diner checks, 50s diner nostalgia
- bowling league as the main structure
- vaporwave grids, synthwave sunsets, neon overload
- Instagram, Strava, or generic social feed hierarchy
- sterile SaaS cards and purple-gradient mobile UI
- fake official bureaucracy

`Evidence` remains the canonical domain term in `CONTEXT.md`; the blacklist is about visual and copy motifs, not the backend/domain noun.

## Open Follow-Ups

The design-language decision exposed a product-language mismatch. These are intentionally backlogged instead of resolved in this note:

- #308: Spike: review Judgment domain language
- #309: Revise product vision away from faux-official judging
- #310: Define judging flow voice and microcopy
- #311: Document brand motif whitelist and blacklist
- #312: Audit user-facing Judge/Judgment language
