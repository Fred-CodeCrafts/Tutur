CREATE TABLE votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id       UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    contributor_id  UUID NOT NULL REFERENCES users(id),
    vote_type       VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote','downvote')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, contributor_id)
);

CREATE TABLE flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id   UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    reason      VARCHAR(50) NOT NULL CHECK (reason IN (
                    'inaccurate_translation','inappropriate_content','duplicate'
                )),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, user_id)
);

CREATE TABLE audio_votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id       UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    contributor_id  UUID NOT NULL REFERENCES users(id),
    vote_type       VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote','downvote')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, contributor_id)
);

CREATE TABLE script_votes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phrase_id       UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    contributor_id  UUID NOT NULL REFERENCES users(id),
    vote_type       VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote','downvote')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (phrase_id, contributor_id)
);
