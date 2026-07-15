CREATE TABLE assets (
    asset_id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    external_id   TEXT NOT NULL UNIQUE,
    symbol        TEXT NOT NULL,
    display_name  TEXT NOT NULL,
    initial_price NUMERIC NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
