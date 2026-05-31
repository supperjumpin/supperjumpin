-- Issue #97: Secondary Terminology Sweep
-- Difficulty -> Commitment
-- Documentation -> Presentation
-- Stunt -> Jump (constraint/index names, disputes column)

ALTER TABLE judgments RENAME COLUMN difficulty TO commitment;

-- Fix constraints in the jumps table
ALTER TABLE jumps DROP CONSTRAINT IF EXISTS jumps_status_check;
ALTER TABLE jumps DROP CONSTRAINT IF EXISTS jumps_check;

ALTER TABLE jumps ADD CONSTRAINT jumps_status_check 
    CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Disqualified Jump', 'Removed Jump'));

ALTER TABLE jumps ADD CONSTRAINT jumps_check 
    CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump', 'Judged Jump', 'Unwitnessed Jump', 'Disqualified Jump', 'Removed Jump')));

-- Fix 'Judged Stunt' -> 'Judged Jump' in final_score check
ALTER TABLE jumps DROP CONSTRAINT stunts_final_score_matches_status;
ALTER TABLE jumps ADD CONSTRAINT jumps_final_score_matches_status CHECK (
    (status = 'Judged Jump' AND final_score IS NOT NULL) OR
    (status != 'Judged Jump' AND final_score IS NULL)
);

-- Rename disputes.stunt_id -> disputes.jump_id (missed in migration 0012)
ALTER TABLE disputes RENAME COLUMN stunt_id TO jump_id;
ALTER INDEX disputes_stunt_id_idx RENAME TO disputes_jump_id_idx;

-- Rename stale index/constraint names from 'stunts' era
ALTER INDEX stunts_pkey RENAME TO jumps_pkey;
ALTER INDEX stunts_group_id_idx RENAME TO jumps_group_id_idx;
ALTER INDEX stunts_player_id_idx RENAME TO jumps_player_id_idx;
ALTER INDEX stunts_season_id_idx RENAME TO jumps_season_id_idx;

ALTER TABLE jumps RENAME CONSTRAINT stunts_group_id_fkey TO jumps_group_id_fkey;
ALTER TABLE jumps RENAME CONSTRAINT stunts_player_id_fkey TO jumps_player_id_fkey;
ALTER TABLE jumps RENAME CONSTRAINT stunts_season_id_fkey TO jumps_season_id_fkey;

-- Rename stale FK constraints on child tables
ALTER TABLE disputes RENAME CONSTRAINT disputes_stunt_id_fkey TO disputes_jump_id_fkey;
ALTER TABLE evidence_upload_authorizations RENAME CONSTRAINT evidence_upload_authorizations_stunt_id_fkey TO evidence_upload_authorizations_jump_id_fkey;
ALTER TABLE evidences RENAME CONSTRAINT evidences_stunt_id_fkey TO evidences_jump_id_fkey;
ALTER TABLE judgments RENAME CONSTRAINT judgments_stunt_id_fkey TO judgments_jump_id_fkey;

-- Rename stale indexes and unique constraints
ALTER INDEX judgments_stunt_id_idx RENAME TO judgments_jump_id_idx;
ALTER INDEX evidence_upload_authorizations_stunt_id_idx RENAME TO evidence_upload_authorizations_jump_id_idx;
ALTER INDEX evidences_stunt_id_key RENAME TO evidences_jump_id_key;
ALTER TABLE judgments RENAME CONSTRAINT judgments_stunt_id_player_id_key TO judgments_jump_id_player_id_key;

-- Rename check constraints on judgments
ALTER TABLE judgments RENAME CONSTRAINT judgments_difficulty_check TO judgments_commitment_check;
ALTER TABLE judgments RENAME CONSTRAINT judgments_documentation_check TO judgments_presentation_check;