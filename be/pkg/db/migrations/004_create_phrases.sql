CREATE TABLE phrases (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text_latin              VARCHAR(500) NOT NULL,
    text_native_script      TEXT,
    script_type             VARCHAR(50) CHECK (script_type IN (
                                'latin','javanese','sundanese',
                                'balinese','lontara','batak','other'
                            )),
    translation             VARCHAR(500) NOT NULL,
    language_code           VARCHAR(20) NOT NULL REFERENCES languages(code),
    tone                    VARCHAR(50) CHECK (tone IN ('formal','netral','kasar')),
    status                  VARCHAR(50) NOT NULL DEFAULT 'pending'
                                CHECK (status IN ('pending','approved','rejected','flagged','ai_failed')),
    script_status           VARCHAR(50) NOT NULL DEFAULT 'none'
                                CHECK (script_status IN ('none','pending','approved','rejected')),
    contributor_id          UUID NOT NULL REFERENCES users(id),
    cultural_context_id     UUID REFERENCES cultural_contexts(id),
    audio_url               TEXT,
    audio_duration_seconds  FLOAT,
    audio_status            VARCHAR(50) NOT NULL DEFAULT 'none'
                                CHECK (audio_status IN ('none','pending','audio_approved','audio_rejected')),
    native_script_image_url TEXT,
    moderated_by            UUID REFERENCES users(id),
    moderated_at            TIMESTAMPTZ,
    upvote_count            INTEGER NOT NULL DEFAULT 0,
    downvote_count          INTEGER NOT NULL DEFAULT 0,
    flag_count              INTEGER NOT NULL DEFAULT 0,
    audio_upvote_count      INTEGER NOT NULL DEFAULT 0,
    audio_downvote_count    INTEGER NOT NULL DEFAULT 0,
    script_upvote_count     INTEGER NOT NULL DEFAULT 0,
    script_downvote_count   INTEGER NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_phrases_language_status ON phrases(language_code, status);
CREATE INDEX idx_phrases_contributor ON phrases(contributor_id);
CREATE INDEX idx_phrases_status ON phrases(status);
