CREATE TABLE IF NOT EXISTS tickets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    client_id BIGINT NOT NULL,
    client_username VARCHAR(255) NOT NULL,
    thread_id INTEGER DEFAULT 0,
    assigned_to_id BIGINT DEFAULT 0,
    assigned_to_username VARCHAR(255) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tickets_client_id
    ON tickets(client_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_open_ticket_per_client
    ON tickets(client_id)
    WHERE closed_at IS NULL;