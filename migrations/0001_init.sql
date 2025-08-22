CREATE TYPE event_status AS ENUM ('received', 'processing', 'processed', 'failed');

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    provider TEXT NOT NULL,
    event_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    status event_status NOT NULL DEFAULT 'received',
    last_error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    UNIQUE (provider, event_id)
);

CREATE INDEX IF NOT EXISTS idx_events_status_created ON events (status, created_at);

CREATE OR REPLACE FUNCTION trg_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER events_updated_at
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE PROCEDURE trg_events_updated_at();