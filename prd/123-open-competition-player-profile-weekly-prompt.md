# PRD: Open Competition, Player Profile, and Weekly Prompt

**Parent Issue:** #68  
**Depends on:** #118, #119, #120, #121, #122  
**Source Design Packages:** #50, #64, #65, #66, #106, #107

## Problem Statement

The accepted MVP uses the **Open** as the v1 competitive context and treats **Weekly Prompt** as a lightweight creation nudge. **Player Profile** gives attribution and history a destination. These surfaces are not required for the first playable Jump/Judgment loop, but they are part of the full MVP because they provide payoff, context, and content rhythm without reintroducing Group Seasons.

They need a parent PRD separate from the core loop because Open scoring, Standings snapshots, Player Profile read models, and Prompt feed placement depend on earlier Jump, Judging, auth, and read-path foundations.

## Solution

Build the MVP's competition and identity surfaces:

1. **Open Month Model & Automatic Assignment** — Every eligible Jump is automatically associated with the current Open month.
2. **Open Soft-Close Computation** — At month-end, Open Final Scores are computed from qualifying Judgments.
3. **Open Standings API & Screen** — Ranked Players for a given month.
4. **Player Profile API & Screen** — Player identity, bio, and Jump history.
5. **Weekly Prompt Pinned Feed Card** — A lightweight creation nudge in the feed.

The Open should act as a **monthly signal surface and lightweight competitive payoff**, not a permanent leaderboard or Group Season substitute.

## User Stories

1. **As a Player**, I want my eligible Jumps to participate in the current Open automatically, so that I do not need to join a Group or Season to compete.
2. **As a Player**, I want the app to know the current Open month, so that Open context is consistent across Judgments and Standings.
3. **As a Judge**, I want my Judgment to be associated with the correct Open month when applicable, so that Open Final Scores are computed fairly.
4. **As a Player**, I want Open Final Scores computed at month-end from qualifying Judgments, so that the monthly competition has a clear payoff moment.
5. **As a Player**, I want Season-provenance Judgments excluded from Open Final Scores in the future, so that competitive contexts remain separate.
6. **As a Player**, I want Open Standings to show ranked Players for a month, so that I can see how the monthly competition resolved.
7. **As a Player**, I want Open Standings to include my rank and score even if I am not in the top N, so that I can see my personal performance.
8. **As a Player**, I want a Player Profile screen that shows my username, bio, avatar, and recent Jumps, so that I have a persistent identity.
9. **As a Player**, I want to view other Players' Profiles, so that I can discover creators and their work.
10. **As a Player**, I want a Weekly Prompt pinned at the top of my feed, so that I have a lightweight creation nudge each week.

## Technical Design

### 1. Open Month Model

- **Schema:** `OpenMonth { id, year, month, startDate, endDate, status: 'active' | 'closed' | 'scored' }`
- **Auto-creation:** A cron job or scheduled function creates a new OpenMonth record on the 1st of each month at 00:00 UTC.
- **Current Open:** A helper function `getCurrentOpenMonth()` returns the active OpenMonth for the current time.

### 2. Automatic Open Assignment

- **When a Jump is created:** If the Jump is eligible (not a draft, not part of a Season), assign it to the current OpenMonth.
- **Eligibility rules:**
  - Jump must be published (status = 'published')
  - Jump must not have a `seasonId` set
  - Jump must be created within the OpenMonth's date range

### 3. Open Soft-Close & Scoring

- **Trigger:** At month-end (last day of month 23:59:59 UTC), a scheduled job runs.
- **Process:**
  1. Mark OpenMonth as 'closed'
  2. Collect all Judgments for Jumps in this OpenMonth
  3. Exclude Judgments that have a `seasonId` (future-proofing)
  4. Compute Final Score per Jump using the existing scoring algorithm
  5. Aggregate scores per Player (sum of all their Jumps' Final Scores in this Open)
  6. Store results in `OpenScore { openMonthId, playerId, totalScore, rank }`
  7. Mark OpenMonth as 'scored'

### 4. Open Standings API

- **Endpoint:** `GET /api/open/standings?month={year}-{month}`
- **Response:**
```json
{
  "month": "2024-03",
  "standings": [
    { "rank": 1, "playerId": "uuid", "username": "player1", "totalScore": 95.5, "avatarUrl": "..." },
    { "rank": 2, "playerId": "uuid", "username": "player2", "totalScore": 88.2, "avatarUrl": "..." }
  ],
  "currentPlayer": { "rank": 15, "totalScore": 42.1 } // if authenticated
}
```
- **Pagination:** Support `limit` and `offset` query params
- **Default:** If no month specified, return current or most recent closed month

### 5. Open Standings Screen

- **Route:** `/open/standings` or accessible from main navigation
- **Components:**
  - Month selector (dropdown or arrows to navigate months)
  - Ranked list with Player avatar, username, score
  - Current Player's position highlighted
  - Loading and empty states

### 6. Player Profile API

- **Endpoint:** `GET /api/players/:playerId`
- **Response:**
```json
{
  "playerId": "uuid",
  "username": "player1",
  "bio": "I make Jumps!",
  "avatarUrl": "...",
  "joinedAt": "2024-01-15T00:00:00Z",
  "recentJumps": [
    { "id": "uuid", "title": "My Jump", "thumbnailUrl": "...", "createdAt": "...", "finalScore": 85.3 }
  ],
  "stats": {
    "totalJumps": 12,
    "bestScore": 95.5,
    "averageScore": 72.1
  }
}
```
- **Pagination:** Support `limit` and `offset` for recentJumps

### 7. Player Profile Screen

- **Route:** `/players/:playerId`
- **Components:**
  - Avatar, username, bio, join date
  - Stats section (total Jumps, best score, average score)
  - Recent Jumps grid/list with thumbnails
  - Link to view all Jumps by this Player
  - Edit Profile button (if viewing own profile)

### 8. Weekly Prompt Pinned Feed Card

- **Data Model:** `WeeklyPrompt { id, weekStart, theme, description, imageUrl }`
- **Creation:** Admin creates weekly prompts via a management interface or seed data
- **Feed Integration:**
  - At the top of the feed, show a pinned card with the current week's prompt
  - Card includes: theme, description, optional image, and a "Create Jump" CTA
  - Card is dismissible (hide for session)
  - If no prompt for current week, show nothing
- **API:** `GET /api/prompts/current` returns the current week's prompt or null

## Data Flow Diagram

```
[Player creates Jump] → [Assign to current OpenMonth] → [Jump published]
                                                              ↓
[Month ends] → [Cron: close OpenMonth] → [Collect Judgments] → [Compute scores] → [Store OpenScores]
                                                              ↓
[Player views Standings] → [GET /api/open/standings] → [Return ranked scores]

[Player views Profile] → [GET /api/players/:id] → [Return player data + recent Jumps]

[Player opens Feed] → [GET /api/prompts/current] → [Show pinned card if exists]
```

## Acceptance Criteria

1. **Open Month:** A new OpenMonth is created automatically on the 1st of each month.
2. **Auto-assignment:** New eligible Jumps are assigned to the current OpenMonth.
3. **Scoring:** At month-end, Open Final Scores are computed and stored.
4. **Standings API:** Returns correct ranked scores for any month.
5. **Standings Screen:** Displays ranked list with current player highlighted.
6. **Player Profile API:** Returns correct player data with recent Jumps.
7. **Player Profile Screen:** Displays player identity and history.
8. **Weekly Prompt:** Pinned card appears in feed when prompt exists.
9. **Error handling:** All APIs return appropriate errors for missing data.
10. **Performance:** Standings API returns in <200ms for typical data sizes.

## Out of Scope

- Group Seasons (deferred to post-MVP)
- Permanent leaderboards
- Open rewards or prizes
- Social features (following, comments)
- Advanced analytics
- Admin UI for prompt management (seed data or direct DB is fine for MVP)

## Timeline

| Milestone | Dependencies | Estimated Effort |
|-----------|--------------|------------------|
| Open Month model & auto-assignment | #118, #119 | 2 days |
| Open scoring & soft-close | #120, #121 | 3 days |
| Open Standings API & Screen | #122 | 2 days |
| Player Profile API & Screen | #118, #119 | 2 days |
| Weekly Prompt Feed Card | #122 | 1 day |
| Integration & testing | All above | 2 days |

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Scoring computation takes too long | High | Use batch processing, index Judgments by OpenMonth |
| OpenMonth creation fails | Medium | Add monitoring and manual fallback |
| Player Profile data model changes | Medium | Keep Profile read model separate from auth model |
| Weekly Prompt feels intrusive | Low | Make card dismissible, track dismissals |

## Future Considerations

- **Open rewards:** Add badges or tokens for top performers
- **Open streaks:** Track consecutive months of participation
- **Player Profile customization:** More fields, themes, achievements
- **Weekly Prompt voting:** Community votes on best prompt responses
- **Open archive:** Browse historical months
