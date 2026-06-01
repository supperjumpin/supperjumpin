CREATE TABLE jumps (
    id TEXT PRIMARY KEY,
    group_id TEXT REFERENCES groups(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    season_id TEXT REFERENCES seasons(id),
    status TEXT NOT NULL CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump', 'Removed Jump')),
    source TEXT NOT NULL,
    destination TEXT NOT NULL,
    food TEXT NOT NULL,
    final_score INT,
    grace_period_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump', 'Removed Jump'))),
    CONSTRAINT jumps_final_score_matches_status CHECK (
        final_score IS NULL
        OR status = 'Judged Jump'
    )
);

CREATE INDEX jumps_group_id_idx ON jumps(group_id);
CREATE INDEX jumps_player_id_idx ON jumps(player_id);
CREATE INDEX jumps_season_id_idx ON jumps(season_id);
