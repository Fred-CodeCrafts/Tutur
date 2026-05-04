CREATE TABLE languages (
    code        VARCHAR(20) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    region      VARCHAR(255),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
