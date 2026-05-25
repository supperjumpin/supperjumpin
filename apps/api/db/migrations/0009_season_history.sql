CREATE TABLE season_history (
    id TEXT PRIMARY KEY,
    season_id TEXT NOT NULL REFERENCES seasons(id),
    action TEXT NOT NULL,
    actor_player_id TEXT NOT NULL REFERENCES players(id),
    actor_role TEXT NOT NULL,
    override BOOLEAN NOT NULL,
    from_status TEXT NOT NULL CHECK (from_status IN ('Active', 'Judging Grace Period', 'Finalized')),
    to_status TEXT NOT NULL CHECK (to_status IN ('Active', 'Judging Grace Period', 'Finalized')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX season_history_season_id_created_at_idx ON season_history(season_id, created_at);
