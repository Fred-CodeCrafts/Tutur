CREATE TABLE words (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surface_form_latin         VARCHAR(255) NOT NULL,
    surface_form_native_script VARCHAR(255),
    root_form_latin            VARCHAR(255) NOT NULL,
    root_form_native_script    VARCHAR(255),
    script_type                VARCHAR(50) CHECK (script_type IN (
                                   'latin','javanese','sundanese',
                                   'balinese','lontara','batak','other'
                               )),
    part_of_speech             VARCHAR(50),
    language_code              VARCHAR(20) NOT NULL REFERENCES languages(code),
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_words_root_form ON words(root_form_latin, language_code);
CREATE INDEX idx_words_language ON words(language_code);
