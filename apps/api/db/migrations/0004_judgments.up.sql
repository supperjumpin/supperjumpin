CREATE TABLE guest_sessions (
    id TEXT PRIMARY KEY,
    judgment_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE guest_sessions ENABLE ROW LEVEL SECURITY;

CREATE TABLE judgments (
    id TEXT PRIMARY KEY,
    jump_id TEXT NOT NULL REFERENCES jumps(id),
    player_id TEXT REFERENCES players(id),
    guest_session_id TEXT REFERENCES guest_sessions(id),
    provenance TEXT NOT NULL DEFAULT 'public',
    commitment INT NOT NULL CHECK (commitment >= 0 AND commitment <= 10),
    transgression INT NOT NULL CHECK (transgression >= 0 AND transgression <= 10),
    creativity INT NOT NULL CHECK (creativity >= 0 AND creativity <= 10),
    presentation INT NOT NULL CHECK (presentation >= 0 AND presentation <= 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((player_id IS NOT NULL) OR (guest_session_id IS NOT NULL))
);
ALTER TABLE judgments ENABLE ROW LEVEL SECURITY;

CREATE UNIQUE INDEX judgments_jump_player_unique ON judgments (jump_id, player_id) WHERE player_id IS NOT NULL;
CREATE UNIQUE INDEX judgments_jump_guest_unique ON judgments (jump_id, guest_session_id) WHERE guest_session_id IS NOT NULL;

CREATE INDEX judgments_jump_id_idx ON judgments(jump_id);
CREATE INDEX judgments_player_id_idx ON judgments(player_id);
