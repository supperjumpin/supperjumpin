ALTER TABLE stunts
    DROP CONSTRAINT stunts_status_check,
    ADD CONSTRAINT stunts_status_check CHECK (status IN ('Idea', 'Planned Stunt', 'Performed Stunt', 'Judged Stunt', 'Unjudged Stunt', 'Disqualified Stunt', 'Removed Stunt')),
    DROP CONSTRAINT stunts_check,
    ADD CONSTRAINT stunts_check CHECK ((season_id IS NULL) OR (status IN ('Planned Stunt', 'Performed Stunt', 'Judged Stunt', 'Unjudged Stunt', 'Disqualified Stunt', 'Removed Stunt'))),
    DROP CONSTRAINT stunts_final_score_matches_status,
    ADD CONSTRAINT stunts_final_score_matches_status CHECK (
        (status = 'Judged Stunt' AND final_score IS NOT NULL)
        OR (status <> 'Judged Stunt' AND final_score IS NULL)
    );
