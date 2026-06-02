CREATE TABLE invites (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id),
    token TEXT NOT NULL UNIQUE,
    created_by_player_id TEXT NOT NULL REFERENCES players(id),
    used_by_player_id TEXT REFERENCES players(id),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX invites_group_id_idx ON invites(group_id);
