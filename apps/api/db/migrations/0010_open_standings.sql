CREATE TABLE open_standings (
    id TEXT PRIMARY KEY,
    year INT NOT NULL,
    month INT NOT NULL,
    player_id TEXT NOT NULL REFERENCES players(id),
    score INT NOT NULL,
    judged_jumps INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(year, month, player_id)
);

CREATE INDEX open_standings_year_month_idx ON open_standings(year, month);
CREATE INDEX open_standings_player_id_idx ON open_standings(player_id);
