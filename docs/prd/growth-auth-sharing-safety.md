# PRD: Growth, Auth, Sharing, and Safety

**Parent Issue:** #68  
**Depends on:** #118, #119, #120, #121  
**Source Design Package:** #50, #64, #65, #66, #106, #107  
**Tracking Issue:** #122

## Problem Statement

The MVP's growth loop depends on Jumps moving outside the app and returning recipients directly into the Judgment loop. At the same time, public user-generated content creates safety obligations: Report flow, team removal, and tombstoning are P0 before public launch. Auth conversion, Guest-to-Player migration, share/deep-link behavior, and moderation cannot be left as loose edge cases because they connect directly to the North Star metric and public visibility risk.

This work needs a dedicated PRD because it crosses mobile auth, backend identity migration, native sharing, deep links, web/share-preview metadata, Report UX, and admin/team removal behavior.

## Solution

Build the growth/auth/safety layer around the core Jump loop: Auth Gate and Display Name Setup, Guest-to-Player Judgment migration, share links and deep links to Jump Detail, share preview metadata/cards, Report submission, manual team/admin removal, and Removed Jump tombstoning across feed, detail, and share previews.

## User Stories

1. **Guest-to-Player Account Creation** – As a Guest Judge, I want to create an Account after Judging, so that I can save my Judgments and become a Player.
2. **Guest Judgment Migration** – As a Guest Judge, I want my existing Guest Judgments migrated when I create an Account, so that my contribution history is not lost.
3. **Auth Gate at Soft Cap** – As a Guest Judge at the soft cap, I want a clear Auth Gate explaining what I unlock, so that I understand why Account creation is being suggested.
4. **Display Name Setup** – As a new Player, I want to set a display name during onboarding, so that my public Jumps and Judgments have understandable attribution.
5. **Share from Feed/Detail** – As a Player, I want to share a Jump from Feed or Detail, so that I can bring friends into the Judgment loop.
6. **Deep Link to Jump Detail** – As a recipient of a shared Jump, I want the link to open directly to Jump Detail, so that I can Judge without navigating the Feed first.
7. **Web Share Preview** – As a recipient without the app installed, I want a useful web preview with Jump metadata, so that I understand what I'm clicking.
8. **Report Content** – As a Player, I want to Report a Jump or Judgment that violates guidelines, so that the community stays safe.
9. **Admin/Team Removal** – As an admin, I want to remove a Jump or Judgment from the platform, so that harmful content is taken down.
10. **Tombstoning** – As a user encountering a removed Jump, I want to see a tombstone placeholder instead of the content, so that I know it was moderated.

## Functional Requirements

### FR1: Auth Gate & Guest-to-Player Migration

- **FR1.1** Guest Judges can Judge up to a soft cap (e.g., 5 Judgments) without an account.
- **FR1.2** After the soft cap, an Auth Gate is shown explaining what they unlock (save history, display name, share Jumps).
- **FR1.3** Auth Gate offers: Sign up with Email, Sign up with Google, Sign up with Apple.
- **FR1.4** On successful account creation, all Guest Judgments are migrated to the new Player identity.
- **FR1.5** Migration is atomic: either all Guest Judgments are migrated or none (rollback on failure).
- **FR1.6** After migration, the Guest session is upgraded to a Player session.

### FR2: Display Name Setup

- **FR2.1** During first-time onboarding after account creation, the Player is prompted to set a display name.
- **FR2.2** Display name must be 3–20 characters, alphanumeric + underscores, unique.
- **FR2.3** Display name can be changed once every 30 days.
- **FR2.4** Display name is shown on Jumps and Judgments the Player creates.

### FR3: Sharing & Deep Links

- **FR3.1** Each Jump has a Share button on Feed card and Detail view.
- **FR3.2** Share invokes the native share sheet with a deep link URL.
- **FR3.3** Deep link format: `supperjumpin://jump/{jumpId}`
- **FR3.4** Universal link (web fallback): `https://supperjumpin.app/jump/{jumpId}`
- **FR3.5** App handles incoming deep links and navigates directly to Jump Detail.
- **FR3.6** If app not installed, universal link opens web preview.

### FR4: Share Preview Metadata

- **FR4.1** Web preview page at `/jump/{jumpId}` returns HTML with Open Graph and Twitter Card meta tags.
- **FR4.2** Meta tags include: title (Jump prompt), description (top Judgment excerpt), image (Jump media thumbnail), author display name.
- **FR4.3** Preview page is server-side rendered (SSR) or statically generated for crawlers.
- **FR4.4** If Jump is removed, preview shows tombstone message.

### FR5: Report Flow

- **FR5.1** Each Jump and Judgment has a Report button (flag icon).
- **FR5.2** Report flow: select reason category (Spam, Harassment, NSFW, Other), optional free-text detail.
- **FR5.3** On submit, Report is stored with reporter ID, target ID, target type, reason, timestamp.
- **FR5.4** Reporter receives confirmation toast.
- **FR5.5** Reports are visible in admin dashboard (future: auto-moderation).

### FR6: Admin/Team Removal

- **FR6.1** Admin dashboard lists reported content with status (pending, reviewed, actioned).
- **FR6.2** Admin can remove a Jump or Judgment with a reason note.
- **FR6.3** Removal is reversible (soft delete) for 30 days.
- **FR6.4** Removed content is tombstoned everywhere.

### FR7: Tombstoning

- **FR7.1** Removed Jumps show a tombstone card in Feed: "This Jump has been removed."
- **FR7.2** Removed Judgments show: "This Judgment has been removed."
- **FR7.3** Jump Detail for removed Jump shows tombstone with reason (if public).
- **FR7.4** Share preview for removed Jump shows tombstone.
- **FR7.5** Original author sees their content as removed with a note.

## Technical Design

### Backend

- **Auth Service** – Handles sign-up, sign-in, session management (JWT).
- **Migration Service** – Handles Guest-to-Player judgment migration.
- **Share Service** – Generates deep links and serves web preview metadata.
- **Report Service** – CRUD for reports, status transitions.
- **Moderation Service** – Handles content removal, soft delete, tombstone state.
- **Database** – PostgreSQL with tables: `users`, `jumps`, `judgments`, `reports`, `moderation_actions`.

### Frontend (Mobile)

- **Auth Gate Screen** – Shown after soft cap or on explicit sign-up.
- **Onboarding Screen** – Display name setup.
- **Share Integration** – Native share sheet with deep link.
- **Deep Link Handler** – URL scheme and universal link handling.
- **Report UI** – Flag button, reason selection, confirmation.
- **Tombstone UI** – Removed content placeholders.

### Web (Share Preview)

- **Next.js page** at `/jump/[id]` with SSR.
- **Open Graph** and **Twitter Card** meta tags.
- **Tombstone detection** – If removed, show appropriate message.

## Data Models

### User
```typescript
interface User {
  id: string;
  email?: string;
  displayName?: string;
  authProvider: 'email' | 'google' | 'apple' | 'guest';
  guestId?: string;
  createdAt: Date;
  updatedAt: Date;
}
```

### Report
```typescript
interface Report {
  id: string;
  reporterId: string;
  targetId: string;
  targetType: 'jump' | 'judgment';
  reason: 'spam' | 'harassment' | 'nsfw' | 'other';
  detail?: string;
  status: 'pending' | 'reviewed' | 'actioned';
  createdAt: Date;
  updatedAt: Date;
}
```

### ModerationAction
```typescript
interface ModerationAction {
  id: string;
  targetId: string;
  targetType: 'jump' | 'judgment';
  action: 'removed' | 'restored';
  reason?: string;
  adminId: string;
  createdAt: Date;
}
```

## Edge Cases & Error Handling

- **Duplicate migration** – If migration is attempted twice, return success (idempotent).
- **Guest session expiry** – If guest session expires before migration, prompt re-auth.
- **Display name taken** – Show inline validation error.
- **Deep link to removed Jump** – Show tombstone.
- **Share without network** – Queue share action for retry.
- **Report on already removed content** – Allow but note content is already moderated.
- **Admin removal of already removed content** – Idempotent.

## Success Metrics

- **Conversion rate**: % of Guest Judges who create an account after soft cap.
- **Share rate**: % of Jumps shared per active Player per week.
- **Deep link open rate**: % of shared link opens that result in a Judgment.
- **Report-to-removal time**: Median time from Report to admin action.
- **Tombstone accuracy**: 100% of removed content shows tombstone.

## Future Considerations

- Automated moderation using ML.
- Appeal flow for removed content.
- Share analytics (who shared, how many opens).
- Referral tracking for growth loops.
- Block/mute user features.

## Open Questions

1. What is the exact soft cap number? (Proposed: 5 Judgments)
2. Should Guest Judges be able to bypass the Auth Gate? (Proposed: No)
3. Should display name changes be logged? (Proposed: Yes, for moderation)
4. Should Reports be anonymous to the reported user? (Proposed: Yes)

## Revision History

| Date | Author | Changes |
|------|--------|---------|
| 2024-01-15 | System | Initial draft from Issue #122 |
