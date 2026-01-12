CREATE TABLE IF NOT EXISTS idempotent_request_statuses(
    request_status_id INT PRIMARY KEY,
    request_status VARCHAR(16) NOT NULL
);
INSERT INTO idempotent_request_statuses(request_status_id, request_status)
VALUES (0, 'PROCESSING') ON CONFLICT DO NOTHING;
INSERT INTO idempotent_request_statuses(request_status_id, request_status)
VALUES (1, 'COMPLETE') ON CONFLICT DO NOTHING;
INSERT INTO idempotent_request_statuses(request_status_id, request_status)
VALUES (2, 'FAILED') ON CONFLICT DO NOTHING;
CREATE TABLE IF NOT EXISTS idempotent_requests(
    request_id uuid NOT NULL,
    user_id uuid NOT NULL,
    request_status_id INT NOT NULL REFERENCES idempotent_request_statuses(request_status_id),
    cached_response_code INT,
    cached_response JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, request_id)
) WITH (ttl_expiration_expression = 'expires_at');
CREATE UNIQUE INDEX IF NOT EXISTS idx_request_id ON idempotent_requests(request_id);