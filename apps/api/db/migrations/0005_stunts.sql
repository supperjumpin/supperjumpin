CREATE TABLE stunts (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    season_id TEXT REFERENCES seasons(id),
    status TEXT NOT NULL CHECK (status IN ('Idea', 'Planned Stunt')),
    source TEXT NOT NULL,
    destination TEXT NOT NULL,
    food TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((season_id IS NULL) OR (status = 'Planned Stunt'))
);

CREATE INDEX stunts_group_id_idx ON stunts(group_id);
CREATE INDEX stunts_player_id_idx ON stunts(player_id);
CREATE INDEX stunts_season_id_idx ON stunts(season_id);
