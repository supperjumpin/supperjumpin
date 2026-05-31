
-- Issue #97: Secondary Terminology Sweep
-- Difficulty -> Commitment
-- Unjudged -> Unwitnessed

ALTER TABLE judgments RENAME COLUMN difficulty TO commitment;

-- Fix constraints in the jumps table
ALTER TABLE jumps DROP CONSTRAINT IF EXISTS jumps_status_check;
ALTER TABLE jumps DROP CONSTRAINT IF EXISTS jumps_check;

ALTER TABLE jumps ADD CONSTRAINT jumps_status_check 
    CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Disqualified Jump', 'Removed Jump'));

ALTER TABLE jumps ADD CONSTRAINT jumps_check 
    CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Disqualified Jump', 'Removed Jump')));
