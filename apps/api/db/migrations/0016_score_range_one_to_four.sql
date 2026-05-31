-- Issue #98: Score Validation Pivot
-- Judgment scores change from 0-10 range to 1-4 forced-choice scale.
-- Drop and recreate CHECK constraints (constraint names already renamed in 0013).

ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_commitment_check;
ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_transgression_check;
ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_creativity_check;
ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_presentation_check;

ALTER TABLE judgments ADD CONSTRAINT judgments_commitment_check
    CHECK (commitment >= 1 AND commitment <= 4);
ALTER TABLE judgments ADD CONSTRAINT judgments_transgression_check
    CHECK (transgression >= 1 AND transgression <= 4);
ALTER TABLE judgments ADD CONSTRAINT judgments_creativity_check
    CHECK (creativity >= 1 AND creativity <= 4);
ALTER TABLE judgments ADD CONSTRAINT judgments_presentation_check
    CHECK (presentation >= 1 AND presentation <= 4);
