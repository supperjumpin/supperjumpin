-- Move stunts to jumps
ALTER TABLE stunts RENAME TO jumps;

-- Rename stunt_id columns to jump_id in all related tables
ALTER TABLE evidence_upload_authorizations RENAME COLUMN stunt_id TO jump_id;
ALTER TABLE evidences RENAME COLUMN stunt_id TO jump_id;
ALTER TABLE judgments RENAME COLUMN stunt_id TO jump_id;

-- Rename documentation to presentation in judgments
ALTER TABLE judgments RENAME COLUMN documentation TO presentation;

-- Update the statuses in the jumps table
UPDATE jumps SET status = 'Planned Jump' WHERE status = 'Planned Stunt';
UPDATE jumps SET status = 'Performed Jump' WHERE status = 'Performed Stunt';

-- Fix the CHECK constraints
-- From 0006_evidence.sql: stunts_status_check and stunts_check
ALTER TABLE jumps DROP CONSTRAINT stunts_status_check;
ALTER TABLE jumps ADD CONSTRAINT jumps_status_check CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump'));

ALTER TABLE jumps DROP CONSTRAINT stunts_check;
ALTER TABLE jumps ADD CONSTRAINT jumps_check CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump')));
