# Judgment Interaction Model (supersedes ADR-0017)

⚠️ **Moot per ADR-0035 (Judgment deleted).** Reactions/Stamps (ADR-0034) replace the Judging interaction. Preserved as historical record.

ADR-0017 confirmed gesture-driven scoring (PanResponder swipes pre-populating all four factors) as the Judging UX. That decision is superseded. The gesture model was designed around an earlier scoring rubric (Difficulty, Transgression, Creativity, Documentation) before the factors were finalized and before the public-feed-first audience was locked.

The confirmed decision: Judging is a single screen showing all four scoring factors simultaneously, with tap-to-select inputs and a 1–4 forced-choice scale per factor. No gestures, no sliders, no sequential per-factor screens. After selecting all four verdicts the Judge proceeds to a confirmation screen ("Enter Judgment into Record") that displays the full ruling before submission. No celebration on confirm — the confirmation reads as a filing receipt.

The gesture model was rejected because: (1) the assumed Judges are mostly strangers on a public feed with low investment, making sequential screens a meaningful drop-off risk; (2) the 1–4 named-tier scale makes tap-to-select the natural interaction — there is no continuous value to drag toward.

Sequential per-factor screens remain worth revisiting if Group Season judging becomes a dominant use case, where Judge investment is higher.

## Scoring factors and tier labels

The four factors replace Difficulty, Transgression, Creativity, and Documentation from the prototype. Difficulty is replaced by Commitment; Documentation was previously renamed Presentation in ADR-0020.

Each factor uses a 1–4 forced-choice scale with named tier labels. No midpoint — Judges must commit to a lean. The labels are in the game's deadpan-institutional register: they describe the Jump's character, not the performer's worth, which mitigates social inhibition on low scores.

**Transgression** — how strongly the Jump violates an expected food/place boundary

| Tier | Label |
|------|-------|
| 4 | Violation is clear, intentional, and defensible under no conventional logic |
| 3 | Violation is evident; some ambient explanation exists but would require effort |
| 2 | Unusual, but a reasonable person might have ended up here by accident |
| 1 | The panel finds insufficient transgression to warrant elevated consideration |

**Creativity** — how novel, thematic, poetic, or absurdly elegant the Jump is

| Tier | Label |
|------|-------|
| 4 | The Jump reveals a structural connection between Source and Destination the panel had not previously considered |
| 3 | The connection is visible; a reasonable person would not have made it by accident |
| 2 | Competent; the pairing is arbitrary rather than discovered |
| 1 | The panel notes a Jump occurred |

**Presentation** — how compellingly the Evidence captures the Jump as a performance

| Tier | Label |
|------|-------|
| 4 | Evidence is self-sufficient; the case stands without the Caption |
| 3 | Evidence supports the Caption; reading order is irrelevant |
| 2 | Evidence is present; the Caption carries significant documentary weight |
| 1 | Evidence is present; it does not advance the case |

**Commitment** — how completely the performer sold the bit with a straight face

| Tier | Label |
|------|-------|
| 4 | The performer's demeanor gives no indication an absurdity is occurring |
| 3 | Commitment holds; a single lapse does not materially undermine the record |
| 2 | Partial; the performer's awareness of the absurdity is visible in the Evidence |
| 1 | Evidence suggests the performer found this funnier than the panel is prepared to |

## Factor display order

Transgression and Creativity appear above Presentation in the layout. Presentation is the most visually anchored factor and will create halo contamination on the factors that follow it if placed first.

## Copy rotation (deferred)

The tier labels are currently static strings. At volume, seeing the same label repeatedly deflates the deadpan effect. The intent is to introduce multiple label variants per tier, rotated so the same verdict reads differently across sessions. This is a Day 2 feature — not MVP scope.
