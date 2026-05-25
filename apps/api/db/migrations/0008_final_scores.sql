ALTER TABLE stunts
    ADD COLUMN final_score INT,
    DROP CONSTRAINT stunts_status_check,
    ADD CONSTRAINT stunts_status_check CHECK (status IN ('Idea', 'Planned Stunt', 'Performed Stunt', 'Judged Stunt', 'Unjudged Stunt')),
    DROP CONSTRAINT stunts_check,
    ADD CONSTRAINT stunts_check CHECK ((season_id IS NULL) OR (status IN ('Planned Stunt', 'Performed Stunt', 'Judged Stunt', 'Unjudged Stunt'))),
    ADD CONSTRAINT stunts_final_score_matches_status CHECK (
        (status = 'Judged Stunt' AND final_score IS NOT NULL)
        OR (status <> 'Judged Stunt' AND final_score IS NULL)
    );
