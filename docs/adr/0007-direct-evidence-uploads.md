# Direct Evidence Uploads

Evidence media will upload directly from the mobile app to object storage using backend-authorized upload paths or signed URLs, rather than passing media through the Go API process. The backend remains authoritative for whether a Player may submit Evidence and for final Evidence records, while storage handles large photo and video payloads.
