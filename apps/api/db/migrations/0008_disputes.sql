CREATE TABLE disputes (
    id TEXT PRIMARY KEY,
    jump_id TEXT NOT NULL REFERENCES jumps(id),
    raised_by_player_id TEXT NOT NULL REFERENCES players(id),
    concern TEXT NOT NULL CHECK (concern IN ('House Rules', 'Credibility', 'Source', 'Destination', 'Food', 'duplicate', 'other')),
    details TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('Open', 'Resolved', 'Overridden')),
    resolution TEXT,
    resolution_reason TEXT,
    resolved_by_player_id TEXT REFERENCES players(id),
    override_resolution TEXT,
    override_reason TEXT,
    override_by_player_id TEXT REFERENCES players(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX disputes_jump_id_idx ON disputes(jump_id);
