ALTER TABLE jumps DROP CONSTRAINT stunts_status_check;
ALTER TABLE jumps
    ADD CONSTRAINT stunts_status_check CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump'));

ALTER TABLE jumps DROP CONSTRAINT stunts_check;
ALTER TABLE jumps
    ADD CONSTRAINT stunts_check CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump')));

CREATE TABLE evidence_upload_authorizations (
    id TEXT PRIMARY KEY,
    jump_id TEXT NOT NULL REFERENCES jumps(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    content_type TEXT NOT NULL,
    media_object_key TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX evidence_upload_authorizations_jump_id_idx ON evidence_upload_authorizations(jump_id);

CREATE TABLE evidences (
    id TEXT PRIMARY KEY,
    jump_id TEXT NOT NULL UNIQUE REFERENCES jumps(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    upload_authorization_id TEXT NOT NULL UNIQUE REFERENCES evidence_upload_authorizations(id),
    caption TEXT NOT NULL,
    media_object_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
