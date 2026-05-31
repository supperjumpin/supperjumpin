CREATE TABLE judgments (
    id TEXT PRIMARY KEY,
    jump_id TEXT NOT NULL REFERENCES jumps(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    difficulty INT NOT NULL CHECK (difficulty >= 0 AND difficulty <= 10),
    transgression INT NOT NULL CHECK (transgression >= 0 AND transgression <= 10),
    creativity INT NOT NULL CHECK (creativity >= 0 AND creativity <= 10),
    presentation INT NOT NULL CHECK (presentation >= 0 AND presentation <= 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (jump_id, player_id)
);

CREATE INDEX judgments_jump_id_idx ON judgments(jump_id);
CREATE INDEX judgments_player_id_idx ON judgments(player_id);
