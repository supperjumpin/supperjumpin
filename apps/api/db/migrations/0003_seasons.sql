CREATE TABLE seasons (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL REFERENCES groups(id),
    commissioner_player_id TEXT NOT NULL REFERENCES players(id),
    status TEXT NOT NULL CHECK (status IN ('Active', 'Judging Grace Period', 'Finalized')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX seasons_group_id_idx ON seasons(group_id);
CREATE UNIQUE INDEX seasons_one_open_per_group_idx
    ON seasons(group_id)
    WHERE status IN ('Active', 'Judging Grace Period');
