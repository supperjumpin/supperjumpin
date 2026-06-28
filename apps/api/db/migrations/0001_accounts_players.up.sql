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

CREATE TABLE prompt_packs (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE prompt_packs ENABLE ROW LEVEL SECURITY;

CREATE TABLE prompts (
    id TEXT PRIMARY KEY,
    pack_id TEXT NOT NULL REFERENCES prompt_packs(id),
    copy TEXT NOT NULL,
    theme TEXT NOT NULL,
    cost_tier TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE prompts ENABLE ROW LEVEL SECURITY;

CREATE TABLE reveal_timeframes (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    duration_hours INTEGER NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE reveal_timeframes ENABLE ROW LEVEL SECURITY;

CREATE TABLE rounds (
    id TEXT PRIMARY KEY,
    community_id TEXT NOT NULL REFERENCES communities(id),
    prompt_id TEXT NOT NULL REFERENCES prompts(id),
    status TEXT NOT NULL DEFAULT 'active',
    reveal_by TIMESTAMPTZ NOT NULL,
    created_by TEXT NOT NULL REFERENCES players(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE rounds ENABLE ROW LEVEL SECURITY;
CREATE UNIQUE INDEX rounds_one_active_per_community ON rounds (community_id) WHERE status = 'active';

CREATE TABLE commits (
    id TEXT PRIMARY KEY,
    round_id TEXT NOT NULL REFERENCES rounds(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    committed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (round_id, player_id)
);
ALTER TABLE commits ENABLE ROW LEVEL SECURITY;

CREATE TABLE jumps (
    id TEXT PRIMARY KEY,
    round_id TEXT NOT NULL REFERENCES rounds(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    caption TEXT NOT NULL DEFAULT '',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (round_id, player_id)
);
ALTER TABLE jumps ENABLE ROW LEVEL SECURITY;

CREATE TABLE jump_evidence (
    id TEXT PRIMARY KEY,
    jump_id TEXT NOT NULL REFERENCES jumps(id),
    url TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);
ALTER TABLE jump_evidence ENABLE ROW LEVEL SECURITY;

CREATE TABLE stamps (
    id TEXT PRIMARY KEY,
    stance TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    glyph TEXT NOT NULL DEFAULT '',
    copy TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE stamps ENABLE ROW LEVEL SECURITY;

CREATE TABLE reactions (
    id TEXT PRIMARY KEY,
    stamp_id TEXT NOT NULL REFERENCES stamps(id),
    jump_id TEXT NOT NULL REFERENCES jumps(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (stamp_id, jump_id, player_id)
);
ALTER TABLE reactions ENABLE ROW LEVEL SECURITY;

CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    round_id TEXT NOT NULL REFERENCES rounds(id),
    jump_id TEXT REFERENCES jumps(id),
    player_id TEXT NOT NULL REFERENCES players(id),
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;

-- Seed reveal timeframes
INSERT INTO reveal_timeframes (id, label, duration_hours, sort_order) VALUES
    ('reveal_tf_800ba5b17c98', '24 hours', 24, 1),
    ('reveal_tf_b7dc9247648a', '3 days', 72, 2),
    ('reveal_tf_785c3a1ef405', '7 days', 168, 3)
ON CONFLICT (id) DO NOTHING;

-- Seed platform-authored prompt catalog
INSERT INTO prompt_packs (id, display_name, description) VALUES
    ('prompt_pack_8c22a060a7c2', 'Kitchen Classics', 'Zero-cost fridge and pantry bits. Use what you already have.'),
    ('prompt_pack_f7f1ecd68377', 'Low-Stakes Smuggling', 'A few bucks gets you in. Snacks and cheap props.'),
    ('prompt_pack_0c066c1877fa', 'Field Operations', 'Real-world audacity. Worth the trip.')
ON CONFLICT (id) DO NOTHING;

INSERT INTO prompts (id, pack_id, copy, theme, cost_tier) VALUES
    ('prompt_0bdcebfc90f3', 'prompt_pack_8c22a060a7c2', 'Turn whatever is in your fridge into a $48 entrée. Plate it. Light it. Write a pretentious menu description.', 'Fine Dining', 'tier_1'),
    ('prompt_902d6371f66e', 'prompt_pack_8c22a060a7c2', 'One item from a meal you are already eating, photographed somewhere it has absolutely no business being.', 'Wrong Room', 'tier_1'),
    ('prompt_95304630e154', 'prompt_pack_8c22a060a7c2', 'Document a single piece of food as evidence in a serious investigation. Police-report caption required.', 'Hostage Situation', 'tier_1'),
    ('prompt_975cca78fa3f', 'prompt_pack_8c22a060a7c2', 'Arrange any food to look deeply unsettling. No gore, just wrongness. Uncanny valley on a plate.', 'Visually Upsetting', 'tier_1'),
    ('prompt_9552d0b8ad5d', 'prompt_pack_f7f1ecd68377', 'Get a gas-station or vending-machine snack into the fanciest context you can reach today without extra travel.', 'Smuggler Run', 'tier_2'),
    ('prompt_eb95844d184b', 'prompt_pack_f7f1ecd68377', 'Present the cheap version as premium — or vice versa. Generic soda in a wine glass. Gas-station pastry on a doily.', 'Brand Betrayal', 'tier_2'),
    ('prompt_bc6ede2f6a49', 'prompt_pack_f7f1ecd68377', 'Two foods that should never touch. Plate them together. Make a case for why it works.', 'Forbidden Pairing', 'tier_2'),
    ('prompt_10f6348588e2', 'prompt_pack_0c066c1877fa', 'Food from one chain, consumed or presented at a direct competitor. The Costco-dog-goes-fancy energy.', 'Across Enemy Lines', 'tier_3'),
    ('prompt_8ec40aa96ad7', 'prompt_pack_0c066c1877fa', 'Commit to a full multi-course meal in a location that serves zero food. Tablecloth, courses, the works.', 'Pop-Up Restaurant', 'tier_3')
ON CONFLICT (id) DO NOTHING;

-- Seed stamp catalog (stance = stable identity; label/glyph/copy = tunable data)
INSERT INTO stamps (id, stance, label, glyph, copy) VALUES
    ('stamp_5ca49a5c9cb3', 'approval', 'Approve', '✅', 'Yes. This is the one.'),
    ('stamp_c4597ab59a37', 'appetite', 'Appetite', '🍽️', 'Would.'),
    ('stamp_8b7b5c74d6ed', 'chaos', 'Chaos', '🌀', 'This person should not be left unsupervised.'),
    ('stamp_eba3c8a70ce8', 'lore', 'Lore', '📜', 'This will be remembered.'),
    ('stamp_f2e80bf2f62c', 'certification', 'Certified', '🏅', 'Textbook. Frame it.'),
    ('stamp_08bd90c58697', 'affectionate_failure', 'Noble Effort', '💀', 'The spirit was willing.')
ON CONFLICT (id) DO NOTHING;
