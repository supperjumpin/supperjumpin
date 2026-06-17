# Direct Evidence Uploads

Evidence media will upload directly from the mobile app to object storage using backend-authorized upload paths or signed URLs, rather than passing media through the Go API process. The backend remains authoritative for whether a Player may submit Evidence and for final Evidence records, while storage handles large photo and video payloads.

**Scoping note:** In local MVP development, object storage is not required. Evidence upload is specified but deferred until hosted infrastructure is introduced. The local MVP runs without `EXPO_PUBLIC_MEDIA_BASE_URL` — Evidence photo handling is additive when object storage is available.

Addendum, post-MVP evolution: When a Jump transitions to Removed Jump (ADR-0024), Evidence rows and object storage objects are preserved for potential appeal review but excluded from all read queries. Deep links return a tombstone page with no Evidence. Object storage keys are scoped by `jump_id` to maintain the one-Evidence-per-Jump constraint.
