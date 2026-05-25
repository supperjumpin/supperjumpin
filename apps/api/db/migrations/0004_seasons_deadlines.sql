-- Add deadline columns to seasons table
ALTER TABLE seasons 
    ADD COLUMN submission_deadline TIMESTAMPTZ NOT NULL,
    ADD COLUMN judging_deadline TIMESTAMPTZ NOT NULL,
    ADD CONSTRAINT seasons_deadlines_order CHECK (judging_deadline > submission_deadline);

-- Add index for deadline-based queries
CREATE INDEX seasons_submission_deadline_idx ON seasons(submission_deadline);
CREATE INDEX seasons_judging_deadline_idx ON seasons(judging_deadline);
