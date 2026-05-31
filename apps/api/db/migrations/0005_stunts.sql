CREATE TABLE jumps (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    season_id TEXT REFERENCES seasons(id),
    status TEXT NOT NULL CHECK (status IN ('Idea', 'Planned Jump')),
    source TEXT NOT NULL,
    destination TEXT NOT NULL,
    food TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((season_id IS NULL) OR (status = 'Planned Jump'))
);

CREATE INDEX stunts_group_id_idx ON jumps(group_id);
CREATE INDEX stunts_player_id_idx ON jumps(player_id);
CREATE INDEX stunts_season_id_idx ON jumps(season_id);
