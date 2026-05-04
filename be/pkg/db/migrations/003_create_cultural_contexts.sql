CREATE TABLE cultural_contexts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_code    VARCHAR(20) NOT NULL REFERENCES languages(code),
    region           VARCHAR(255),
    usage_situation  VARCHAR(500),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
