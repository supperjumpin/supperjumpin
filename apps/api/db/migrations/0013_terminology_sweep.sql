
-- Issue #97: Secondary Terminology Sweep
-- Difficulty -> Commitment
-- Unjudged -> Unwitnessed

ALTER TABLE judgments RENAME COLUMN difficulty TO commitment;

-- Fix constraints in the stunts table
ALTER TABLE stunts DROP CONSTRAINT IF EXISTS stunts_status_check;
ALTER TABLE stunts DROP CONSTRAINT IF EXISTS stunts_check;

ALTER TABLE stunts ADD CONSTRAINT stunts_status_check 
    CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Disqualified Jump', 'Removed Jump'));

ALTER TABLE stunts ADD CONSTRAINT stunts_check 
    CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Disqualified Jump', 'Removed Jump')));
