CREATE TABLE phrase_words (
    phrase_id   UUID NOT NULL REFERENCES phrases(id) ON DELETE CASCADE,
    word_id     UUID NOT NULL REFERENCES words(id),
    position    INTEGER NOT NULL,
    PRIMARY KEY (phrase_id, word_id)
);
