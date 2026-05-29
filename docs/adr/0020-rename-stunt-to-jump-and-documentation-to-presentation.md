# Rename Stunt → Jump and Documentation → Presentation

Supersedes: ADR-0008 (Stunt language), ADR-0005, ADR-0006, ADR-0009, ADR-0011, ADR-0012, ADR-0016, ADR-0018 (Stunt language in body)

## Decision

The canonical domain term for a playable Supperjumpin attempt is **Jump**, not Stunt. The scoring factor for how compellingly Evidence captures a Jump as a performance is **Presentation**, not Documentation.

`CONTEXT.md` already reflects both changes. Code, OpenAPI schema, generated TypeScript client, SQL schema, and migration files use the old terms and must be updated in a follow-up sweep (see implementation issue).

## Rationale

**Stunt → Jump**: Stunt was the original working term borrowed from the source article. It was flagged early as wrong for the game's spirit (issue #50: "calling them Stunts is dumb"). Jump is the name the game is built around — it appears in the product name, the core mechanic description, and the public vocabulary. Carrying two names (Jump in docs, Stunt in code) creates confusion in every code review, schema discussion, and future ADR. The rename is a one-time cost with permanent clarity payoff. There is no deployed API to preserve compatibility with.

**Documentation → Presentation**: Documentation reads like a file management checklist alongside Difficulty, Transgression, and Creativity. Presentation names the same scoring factor — how compellingly the Evidence frames the Jump as a performance — while fitting the deadpan-institutional register of the rest of the rubric. The distinction from Credibility (did it look good vs. did it happen) sharpens with Presentation as the label.

## Tone guidance

**Jump** is the one term that should feel playful and carry the game's absurdist identity. All other terms use plain, legible language — the humor comes from applying straight-faced institutional vocabulary (Season Commissioner, Standings, Judging Grace Period, House Rules) to intrinsically absurd content. Invented-cute renames would deflate the joke; the deadpan is load-bearing.

Specific calls:
- **Transgression**: keep as-is. Its slight formality is intentional — it names the scoring factor precisely and the formal register is part of the game's character.
- **House Rules**: player-facing, deliberately reads like a kitchen-table game. Keep.
- **Season Commissioner**: one of the best terms in the set. Borrowed from fantasy sports, immediately funny in context. Keep.
- **Standings** over Leaderboard: correct. Leaderboard signals gamification grind; Standings signals a real sporting body adjudicating Crunchwrap placement.

## Historical ADR handling

ADRs 0005–0018 that contain Stunt language are left unedited. They are historical records. Future readers should read Jump wherever those ADRs say Stunt. This ADR is the canonical rename record.

## What needs to change (implementation)

See implementation issue for the full sweep. In summary:

- Go source: all `Stunt`/`stunt` identifiers, type names, error vars, status strings, file names
- OpenAPI schema: `Stunt` schema, `stuntId` path params, `/v1/stunts/` routes, status string values (`"Performed Stunt"`, `"Disqualified Stunt"`, etc.)
- SQL migrations and queries: `stunts` table, `stunt_id` columns, status enum values
- Generated TypeScript client (`generated.d.ts`, `index.d.ts`): regenerate from updated OpenAPI
- Mobile app (`App.tsx`): update imports and variable names
- Scoring factor: `documentation` → `presentation` everywhere the four-factor rubric appears (Go, OpenAPI, TypeScript, SQL)
- Open issue titles containing "Stunt": update during the sweep PR
- ADR-0017 references `documentation` as a scoring factor: update inline since it is the most recent gesture-scoring ADR and still active
