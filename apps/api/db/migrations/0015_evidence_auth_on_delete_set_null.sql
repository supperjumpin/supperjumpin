-- Make upload_authorization_id nullable and ON DELETE SET NULL
-- This allows the server to delete consumed upload authorizations
-- while preserving the evidence record.

ALTER TABLE evidences DROP CONSTRAINT evidences_upload_authorization_id_fkey;
ALTER TABLE evidences ALTER COLUMN upload_authorization_id DROP NOT NULL;
ALTER TABLE evidences ADD CONSTRAINT evidences_upload_authorization_id_fkey
    FOREIGN KEY (upload_authorization_id) REFERENCES evidence_upload_authorizations(id)
    ON DELETE SET NULL;