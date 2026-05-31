ALTER TABLE jumps
    ADD COLUMN final_score INT,
    DROP CONSTRAINT jumps_status_check,
    ADD CONSTRAINT jumps_status_check CHECK (status IN ('Idea', 'Planned Jump', 'Performed Jump', 'Judged Jump', 'Unjudged Jump')),
    DROP CONSTRAINT jumps_check,
    ADD CONSTRAINT jumps_check CHECK ((season_id IS NULL) OR (status IN ('Planned Jump', 'Performed Jump', 'Judged Jump', 'Unjudged Jump'))),
    ADD CONSTRAINT jumps_final_score_matches_status CHECK (
        (status = 'Judged Jump' AND final_score IS NOT NULL)
        OR (status <> 'Judged Jump' AND final_score IS NULL)
    );
