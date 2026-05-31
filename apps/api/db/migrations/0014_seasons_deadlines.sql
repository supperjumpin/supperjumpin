-- Add submission_deadline and judging_deadline to seasons
ALTER TABLE seasons ADD COLUMN IF NOT EXISTS submission_deadline TIMESTAMPTZ NOT NULL DEFAULT now() + interval '7 days';
ALTER TABLE seasons ADD COLUMN IF NOT EXISTS judging_deadline TIMESTAMPTZ NOT NULL DEFAULT now() + interval '14 days';