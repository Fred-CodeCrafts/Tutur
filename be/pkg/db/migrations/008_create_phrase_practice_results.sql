CREATE TABLE phrase_practice_results (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id  UUID NOT NULL REFERENCES users(id),
    phrase_id   UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    result      VARCHAR(20) NOT NULL CHECK (result IN ('tahu','belum_tahu')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_practice_learner ON phrase_practice_results(learner_id, phrase_id);
