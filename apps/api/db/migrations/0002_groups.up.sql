CREATE TABLE groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE group_memberships (
    group_id TEXT NOT NULL REFERENCES groups(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    role TEXT NOT NULL CHECK (role IN ('Group Admin', 'Player')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, player_id)
);

CREATE INDEX group_memberships_player_id_idx ON group_memberships(player_id);
