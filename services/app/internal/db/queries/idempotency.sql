-- name: GetRequest :one
SELECT * FROM idempotent_requests
WHERE request_id = $1;

-- name: CreateRequest :execrows
INSERT INTO idempotent_requests(
    request_id,
    user_id,
    request_status_id,
    expires_at
) VALUES ($1, $2, $3, $4);

-- name: UpdateRequest :execrows
UPDATE idempotent_requests SET
    request_status_id = $1,
    cached_response_code = $2,
    cached_response = $3,
    expires_at = $4
WHERE request_id = $5 AND request_status_id = $6 AND expires_at < $7;

