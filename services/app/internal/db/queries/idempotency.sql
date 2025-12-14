-- name: GetRequest :one
SELECT * FROM idempotent_requests
WHERE request_id = $1;

-- name: CreateRequest :exec
INSERT INTO idempotent_requests(
    request_id,
    user_id,
    request_status_id,
    expires_at
) VALUES ($1, $2, $3, $4);

-- name: UpdateRequestFailed :execrows
UPDATE idempotent_requests SET
    request_status_id = 2
WHERE request_id = $1 AND request_status_id = 0 AND expires_at < NOW();

-- name: UpdateRequestSuccess :execrows
UPDATE idempotent_requests SET
    request_status_id = 1,
    expires_at = $1
WHERE request_id = $2 AND request_status_id = 0 AND expires_at > NOW();

-- name: UpdateRequestReprocess :execrows
UPDATE idempotent_requests SET
    request_status_id = 0,
    expires_at = $1
WHERE request_id = $2 AND
    (request_status_id = 2 OR
    (request_status_id = 0 AND expires_at < NOW()));
