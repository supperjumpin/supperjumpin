CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;

CREATE TABLE auth_identities (
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (provider, subject)
);
ALTER TABLE auth_identities ENABLE ROW LEVEL SECURITY;

CREATE TABLE communities (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE communities ENABLE ROW LEVEL SECURITY;

CREATE TABLE players (
    id TEXT PRIMARY KEY,
    account_id TEXT UNIQUE REFERENCES accounts(id),
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE players ENABLE ROW LEVEL SECURITY;

CREATE TABLE external_identity (
    platform TEXT NOT NULL,
    platform_server_id TEXT NOT NULL,
    platform_user_id TEXT NOT NULL,
    player_id TEXT NOT NULL REFERENCES players(id),
    community_id TEXT NOT NULL REFERENCES communities(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (platform, platform_server_id, platform_user_id)
);
ALTER TABLE external_identity ENABLE ROW LEVEL SECURITY;
