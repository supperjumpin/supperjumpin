-- Issue #98: Score Validation Pivot (DOWN)
-- Revert CHECK constraints back to 0-10 range.
-- Does NOT reverse existing data (data migration is a one-way clamp).

ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_commitment_check;
ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_transgression_check;
ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_creativity_check;
ALTER TABLE judgments DROP CONSTRAINT IF EXISTS judgments_presentation_check;

ALTER TABLE judgments ADD CONSTRAINT judgments_commitment_check
    CHECK (commitment >= 0 AND commitment <= 10);
ALTER TABLE judgments ADD CONSTRAINT judgments_transgression_check
    CHECK (transgression >= 0 AND transgression <= 10);
ALTER TABLE judgments ADD CONSTRAINT judgments_creativity_check
    CHECK (creativity >= 0 AND creativity <= 10);
ALTER TABLE judgments ADD CONSTRAINT judgments_presentation_check
    CHECK (presentation >= 0 AND presentation <= 10);
