CREATE TABLE jumps (
    id TEXT PRIMARY KEY,
    player_id TEXT NOT NULL REFERENCES players(id),
    season_id TEXT,
    status TEXT NOT NULL CHECK (status IN ('Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump', 'Removed Jump')),
    source TEXT NOT NULL,
    destination TEXT NOT NULL,
    food TEXT NOT NULL,
    final_score INT,
    open_final_score INT,
    removed_at TIMESTAMPTZ,
    grace_period_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT jumps_final_score_matches_status CHECK (
        final_score IS NULL
        OR status = 'Judged Jump'
    )
);

CREATE INDEX jumps_player_id_idx ON jumps(player_id);
