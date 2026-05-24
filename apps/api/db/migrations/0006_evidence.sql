ALTER TABLE stunts DROP CONSTRAINT stunts_status_check;
ALTER TABLE stunts
    ADD CONSTRAINT stunts_status_check CHECK (status IN ('Idea', 'Planned Stunt', 'Performed Stunt'));

ALTER TABLE stunts DROP CONSTRAINT stunts_check;
ALTER TABLE stunts
    ADD CONSTRAINT stunts_check CHECK ((season_id IS NULL) OR (status IN ('Planned Stunt', 'Performed Stunt')));

CREATE TABLE evidence_upload_authorizations (
    id TEXT PRIMARY KEY,
    stunt_id TEXT NOT NULL REFERENCES stunts(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    content_type TEXT NOT NULL,
    media_object_key TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX evidence_upload_authorizations_stunt_id_idx ON evidence_upload_authorizations(stunt_id);

CREATE TABLE evidences (
    id TEXT PRIMARY KEY,
    stunt_id TEXT NOT NULL UNIQUE REFERENCES stunts(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    upload_authorization_id TEXT NOT NULL UNIQUE REFERENCES evidence_upload_authorizations(id),
    caption TEXT NOT NULL,
    media_object_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
