-- Issue #98: Score Validation Pivot
-- Judgment scores change from 0-10 range to 1-4 forced-choice scale.
--
-- 1. Migrate existing data: clamp old 0-10 scores into 1-4 range
--    using a linear mapping: new = ROUND(old / 10.0 * 3) + 1
--    This maps 0→1, 5→2, 10→4.
-- 2. Drop old CHECK constraints (constraint names renamed in 0013)
-- 3. Add new CHECK constraints enforcing 1-4 range

UPDATE judgments SET
  commitment = GREATEST(1, LEAST(4, ROUND(commitment::numeric / 10.0 * 3) + 1)),
  transgression = GREATEST(1, LEAST(4, ROUND(transgression::numeric / 10.0 * 3) + 1)),
  creativity = GREATEST(1, LEAST(4, ROUND(creativity::numeric / 10.0 * 3) + 1)),
  presentation = GREATEST(1, LEAST(4, ROUND(presentation::numeric / 10.0 * 3) + 1));

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
