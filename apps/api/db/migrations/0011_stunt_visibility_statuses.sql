ALTER TABLE jumps
    DROP CONSTRAINT stunts_status_check,
    ADD CONSTRAINT stunts_status_check CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump', 'Removed Jump')),
    DROP CONSTRAINT stunts_check,
    ADD CONSTRAINT stunts_check CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump', 'Judged Jump', 'Unjudged Jump', 'Disqualified Jump', 'Removed Jump'))),
    DROP CONSTRAINT stunts_final_score_matches_status,
    ADD CONSTRAINT stunts_final_score_matches_status CHECK (
        (status = 'Judged Jump' AND final_score IS NOT NULL)
        OR (status <> 'Judged Jump' AND final_score IS NULL)
    );
