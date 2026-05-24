CREATE TABLE judgments (
    id TEXT PRIMARY KEY,
    stunt_id TEXT NOT NULL REFERENCES stunts(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    difficulty INT NOT NULL CHECK (difficulty >= 0 AND difficulty <= 10),
    transgression INT NOT NULL CHECK (transgression >= 0 AND transgression <= 10),
    creativity INT NOT NULL CHECK (creativity >= 0 AND creativity <= 10),
    documentation INT NOT NULL CHECK (documentation >= 0 AND documentation <= 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (stunt_id, player_id)
);

CREATE INDEX judgments_stunt_id_idx ON judgments(stunt_id);
CREATE INDEX judgments_player_id_idx ON judgments(player_id);
